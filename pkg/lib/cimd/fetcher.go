package cimd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

const (
	// MaxDocumentBytes is draft-ietf-oauth-client-id-metadata-document-02
	// §8.7's recommended maximum. Enforced progressively while reading the
	// body, never from Content-Length, which a server may omit or misstate.
	MaxDocumentBytes = 5120

	// FetchTimeout covers DNS resolution, TLS handshake and reading the
	// response body, because it is applied as http.Client.Timeout. It
	// matches the existing blocking-webhook per-call default
	// (hook.sync_hook_timeout_seconds) and every other
	// httputil.NewExternalClient call site in this repo, all of which use 5s.
	FetchTimeout = 5 * time.Second

	// RefetchInterval bounds how often a persisted CIMD row is refreshed.
	// Consumed in Part 3; declared here so every fixed CIMD limit is in one
	// place.
	RefetchInterval = 1 * time.Hour
)

// None of the limits above is project-configurable. docs/specs/cimd.md §
// Configuration says so outright: they are not tuning knobs, they are the
// bounds on what an unauthenticated caller can make this server do to a
// third party. A project owner raising MaxDocumentBytes or the timeout
// raises the cost that OTHER people's infrastructure pays.

// SafeDialer, and the error it returns, now live in pkg/util/httputil so that
// the other places fetching a URL taken from configuration share exactly this
// dialer rather than a second implementation of the same rules. The aliases
// keep this package's spelling, and its tests, unchanged.
type SafeDialer = httputil.SafeDialer

type netipResolver = httputil.NetIPResolver

var errBlockedAddress = httputil.ErrBlockedAddress

// Errors returned by Fetch. These are package internals: Fetcher's tests
// assert exactly which rule fired, but nothing outside pkg/lib/cimd ever
// sees one of these directly -- Service.EnsureClientResolved is the one and
// only place that collapses every fetch/validation failure into a single
// outcome (docs/specs/cimd.md § Authgear as an SSRF/Probing Oracle: an
// attacker must not be able to distinguish "connection refused" from
// "timeout" from "valid JSON but failed validation").
var (
	ErrResponseNotOK    = errors.New("cimd: response status is not 2xx")
	ErrResponseTooLarge = errors.New("cimd: response exceeds the maximum document size")
)

var FetcherLogger = slogutil.NewLogger("cimd-fetcher")

// CIMDHTTPClients holds the two clients Fetcher chooses between. A named
// wrapper type rather than a bare *http.Client so wire can distinguish it
// from the other *http.Client providers already in the graph (same pattern
// as SiteAdminHTTPClient).
//
// Two clients, not one: selecting between them keeps the decision in ONE
// greppable place (Fetcher.clientFor), and the permissive transport is a
// distinct object a reviewer can search for.
//
// These are per-project, not process-level, because
// insecure_fetch_address_allowed_hosts is per-project. That costs the
// connection pooling a process-level client would have, which the fetch
// path can afford: it is capped at 10/min/project.
type CIMDHTTPClients struct {
	Strict   *http.Client
	Insecure *http.Client
}

func ProvideCIMDHTTPClients(f *config.HTTPFeatureConfig) *CIMDHTTPClients {
	allowedHosts := f.GetInsecureFetchAddressAllowedHosts()
	return &CIMDHTTPClients{
		Strict:   newCIMDHTTPClient(false, allowedHosts),
		Insecure: newCIMDHTTPClient(true, allowedHosts),
	}
}

// newCIMDHTTPClient builds one of the two http.Clients used for CIMD fetches.
// It is httputil.NewSSRFSafeExternalClient, the same client every other fetch
// of a URL this deployment did not choose uses -- which is where the address
// rules, the no-proxy transport, the no-redirect policy and the refusal log
// all come from.
//
// No redirects: docs/specs/cimd.md § SSRF Protection, "Follow 0 redirects -- a
// redirect target hasn't been through Client ID Format validation". A 3xx is
// returned as a response rather than followed, and Fetch's 2xx check rejects
// it.
func newCIMDHTTPClient(allowNonPublicAddresses bool, allowedHosts []string) *http.Client {
	return httputil.NewSSRFSafeExternalClient(FetchTimeout, httputil.SSRFSafeExternalClientOptions{
		AllowNonPublicAddresses: allowNonPublicAddresses,
		AllowedHosts:            allowedHosts,
		Sink:                    "oauth.client_id_metadata_document",
	})
}

// Fetcher performs the one and only network call CIMD ever makes: an HTTP
// GET against an attacker-chosen client_id URL.
type Fetcher struct {
	HTTPClients *CIMDHTTPClients
	// HTTPFeatureConfig supplies http.insecure_fetch_address_allowed. The
	// host allowlist is applied by the clients themselves.
	HTTPFeatureConfig *config.HTTPFeatureConfig
	// AppID is read only by clientFor's warning log.
	AppID config.AppID
}

// Fetch GETs the document at u and returns its raw bytes. u must already
// have passed oauthclient.ParseCIMDClientID and the caller's
// allowed_domains check. Errors are specific; Service is what collapses
// them.
//
// The response Content-Type is deliberately not checked: spec § Validation
// requires only "MUST be 2xx and MUST parse as a JSON object within the
// size limit". Requiring application/json would reject real static-file
// hosts that serve .json as text/plain while adding nothing -- the body
// still has to parse as a JSON object with a matching client_id.
func (f *Fetcher) Fetch(ctx context.Context, u *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Authgear")

	resp, err := f.clientFor(ctx, u).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// MUST be 2xx. A 3xx lands here too (redirects are not followed) and is
	// refused by this same check.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%w: %d", ErrResponseNotOK, resp.StatusCode)
	}

	// Progressive size enforcement: read at most MaxDocumentBytes+1 and
	// refuse if the extra byte materialised. Content-Length is never
	// consulted -- spec § SSRF Protection requires this explicitly, "since a
	// server can omit or misstate it". Reading +1 rather than exactly the
	// limit is what makes "exactly 5120 bytes" acceptable and "5121 bytes"
	// refused, without a separate length probe.
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxDocumentBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > MaxDocumentBytes {
		return nil, ErrResponseTooLarge
	}

	return body, nil
}

// clientFor selects the strict or the permissive transport for this
// project. This is the ONLY place in the CIMD fetch path that reads
// http.insecure_fetch_address_allowed; everything else takes a client. A
// reviewer auditing "when can Authgear reach a private address" reads this
// function and nothing else.
func (f *Fetcher) clientFor(ctx context.Context, u *url.URL) *http.Client {
	if !f.HTTPFeatureConfig.IsInsecureFetchAddressAllowed() {
		return f.HTTPClients.Strict
	}
	// Without this log, a flag left set on a deployed project is completely
	// invisible. Volume is bounded to <= 10/min/project by the per-project
	// fetch rate limit (Part 4), so this cannot flood a log pipeline, and it
	// is deliberately Warn rather than Info so it surfaces in default log
	// configurations.
	logger := FetcherLogger.GetLogger(ctx)
	logger.Warn(ctx, "cimd: fetching with SSRF address protection disabled",
		slog.String("app_id", string(f.AppID)),
		slog.String("host", u.Hostname()),
		slog.String("flag", httputil.InsecureFetchAddressAllowedFlag),
	)
	return f.HTTPClients.Insecure
}
