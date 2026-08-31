package cimd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"syscall"
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

var errBlockedAddress = errors.New("cimd: address is not publicly routable")

// netipResolver is the one *net.Resolver method SafeDialer needs, pulled
// out as an interface so tests can stub DNS resolution without spinning up
// a real DNS server. *net.Resolver satisfies it as-is.
type netipResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

// SafeDialer implements docs/specs/cimd.md § SSRF Protection's first two
// rules together, because neither works without the other:
//
//   - Resolve the hostname once per fetch attempt and connect only to an
//     address validated in that same resolution -- a second, independent
//     resolution at connect time is vulnerable to DNS rebinding.
//   - Check every address a hostname resolves to, not just the first --
//     reject the whole hostname if any A/AAAA record is non-publicly-
//     routable, rather than filtering to the routable subset. A mixed
//     answer is an attack signature, not a misconfiguration to accommodate.
//
// AllowNonPublicAddresses carries insecure_fetch_address_allowed as a plain
// field rather than a config read, so the dialer stays a pure function of
// its inputs. There is no DevMode field: exactly one mechanism widens this
// policy.
type SafeDialer struct {
	Resolver                netipResolver // nil means net.DefaultResolver
	AllowNonPublicAddresses bool
}

func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}

	// An IP-literal host never goes through DNS. oauthclient.ParseCIMDClientID
	// lets an IP-literal client_id through shape validation; this is where
	// it is refused.
	if literal, err := netip.ParseAddr(host); err == nil {
		if !d.allow(literal) {
			return nil, errBlockedAddress
		}
		return d.dial(ctx, network, netip.AddrPortFrom(literal, uint16(portNum)))
	}

	resolver := d.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addrs, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errBlockedAddress
	}
	for _, a := range addrs {
		if !d.allow(a) {
			return nil, errBlockedAddress
		}
	}

	// Connect to an address from THAT resolution. Each is already validated;
	// trying them in order keeps a dual-stack host with an unreachable AAAA
	// working.
	var lastErr error
	for _, a := range addrs {
		conn, err := d.dial(ctx, network, netip.AddrPortFrom(a, uint16(portNum)))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (d *SafeDialer) allow(addr netip.Addr) bool {
	if httputil.IsPubliclyRoutable(addr) {
		return true
	}
	return d.AllowNonPublicAddresses
}

// dial connects to the already-validated address ap. Control re-validates
// the concrete syscall address as defence in depth: by construction it
// should never fire, and it exists so a future refactor that reintroduces
// hostname-based dialling cannot silently reopen the DNS-rebinding hole.
//
// Only network values "tcp", "tcp4" and "tcp6" occur here.
func (d *SafeDialer) dial(ctx context.Context, network string, ap netip.AddrPort) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: FetchTimeout,
		Control: func(network, address string, c syscall.RawConn) error {
			parsed, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			if !d.allow(parsed.Addr()) {
				return errBlockedAddress
			}
			return nil
		},
	}
	return dialer.DialContext(ctx, network, ap.String())
}

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

// CIMDHTTPClients holds the two process-level, connection-pooled clients
// Fetcher chooses between. A named wrapper type rather than a bare
// *http.Client so wire can distinguish it from the other *http.Client
// providers already in the graph (same pattern as SiteAdminHTTPClient).
//
// Two clients, not one: the address policy is a property of the transport,
// and the transport is process-level (connection pooling). Selecting
// between two named clients keeps the decision in ONE greppable place
// (Fetcher.clientFor) instead of threading a boolean into the dial path on
// every call, and it means the permissive transport is a distinct object a
// reviewer can search for.
type CIMDHTTPClients struct {
	Strict   *http.Client
	Insecure *http.Client
}

func ProvideCIMDHTTPClients() *CIMDHTTPClients {
	return &CIMDHTTPClients{
		Strict:   newCIMDHTTPClient(false),
		Insecure: newCIMDHTTPClient(true),
	}
}

// newCIMDHTTPClient builds one of the two http.Clients used for CIMD
// fetches. It reuses httputil.NewExternalClientWithOptions so the fetch is
// otel-instrumented and non-redirect-following on the same code path as
// every other outbound call in the repo, and supplies the SSRF-safe
// transport underneath.
//
// FollowRedirect: false gives CheckRedirect = ErrUseLastResponse, so a 3xx
// is returned as a response rather than followed; Fetch's 2xx check then
// rejects it. docs/specs/cimd.md § SSRF Protection: "Follow 0 redirects --
// a redirect target hasn't been through Client ID Format validation, and
// would otherwise let the previous two rules be bypassed."
func newCIMDHTTPClient(allowNonPublicAddresses bool) *http.Client {
	dialer := &SafeDialer{AllowNonPublicAddresses: allowNonPublicAddresses}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   FetchTimeout,
		ResponseHeaderTimeout: FetchTimeout,
		// No Proxy. http.ProxyFromEnvironment would route the request
		// through a proxy chosen by the environment, and the proxy -- not
		// this dialer -- would then do the name resolution, silently
		// bypassing every rule above.
		Proxy: nil,
	}
	return httputil.NewExternalClientWithOptions(FetchTimeout, httputil.ExternalClientOptions{
		FollowRedirect: false,
		Transport:      transport,
	})
}

// Fetcher performs the one and only network call CIMD ever makes: an HTTP
// GET against an attacker-chosen client_id URL.
type Fetcher struct {
	HTTPClients *CIMDHTTPClients
	// OAuthFeatureConfig supplies insecure_fetch_address_allowed. Already
	// fanned out by wire (pkg/lib/deps/deps_config.go:74).
	OAuthFeatureConfig *config.OAuthFeatureConfig
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
// insecure_fetch_address_allowed; everything else takes a client. A
// reviewer auditing "when can Authgear reach a private address" reads this
// function and nothing else.
func (f *Fetcher) clientFor(ctx context.Context, u *url.URL) *http.Client {
	if !f.OAuthFeatureConfig.GetClientIDMetadataDocument().IsInsecureFetchAddressAllowed() {
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
		slog.String("flag", "oauth.client_id_metadata_document.insecure_fetch_address_allowed"),
	)
	return f.HTTPClients.Insecure
}
