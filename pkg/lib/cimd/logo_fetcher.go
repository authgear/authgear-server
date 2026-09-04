package cimd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	// MaxLogoBytes bounds both the fetch and the cache entry. 256 KiB is
	// generous for a consent-screen logo (the document itself is capped at
	// MaxDocumentBytes) and small enough that a full cache of them is a
	// bounded, sane Redis footprint.
	MaxLogoBytes = 256 * 1024
)

var (
	ErrLogoInvalidURI      = errors.New("cimd: invalid logo_uri")
	ErrLogoContentType     = errors.New("cimd: unsupported logo content type")
	ErrLogoContentMismatch = errors.New("cimd: logo content does not match an allowed image type")
	ErrLogoTooLarge        = errors.New("cimd: logo exceeds the maximum size")
)

// allowedLogoContentTypes deliberately excludes image/svg+xml: an SVG is a
// document, not an image -- it can contain <script>, external references and
// CSS, and it would be attacker-supplied content served from Authgear's own
// origin. The client_logo endpoint's CSP + sandbox headers defend against
// this too, but not accepting the format at all is defence in depth. Also
// excludes image/x-icon and image/vnd.microsoft.icon -- no reason to accept
// a format nothing needs.
var allowedLogoContentTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
	"image/webp": true,
}

// LogoFetcher reuses the document fetch's clients and its client-selection
// rule verbatim -- same SafeDialer, same strict/insecure pair (see
// Fetcher.clientFor), same insecure_fetch_address_allowed feature flag. A
// logo fetch that used the strict client while the document fetch used the
// permissive one would make a project's own logo unfetchable in exactly the
// environments the flag exists for; the reverse would be a policy hole. It
// is a sibling of Fetcher, not a parameterisation of it, so neither can
// drift into the other's limits (MaxDocumentBytes vs MaxLogoBytes, the
// application/json Accept header vs the image one).
type LogoFetcher struct {
	HTTPClients *CIMDHTTPClients
	// OAuthFeatureConfig supplies insecure_http_allowed and
	// insecure_fetch_address_allowed.
	OAuthFeatureConfig *config.OAuthFeatureConfig
	// AppID is read only by clientFor's warning log.
	AppID config.AppID
}

// Fetch GETs logoURI and returns its bytes and the SNIFFED content type
// (never the server-declared one -- see the sniff check below). logoURI
// comes from an already-resolved client's already-validated document, but
// this function re-validates the scheme anyway: it is what turns a string
// into an outbound request, and the check belongs where the request is
// made, not where the caller trusts it was made once already.
func (f *LogoFetcher) Fetch(ctx context.Context, logoURI string) (body []byte, contentType string, err error) {
	u, parseErr := url.Parse(logoURI)
	// https, unless insecure_http_allowed is set for this project -- the
	// same flag and the same posture that lets the document itself declare
	// an http logo_uri (document.go rule 8). The two MUST agree: if
	// validation stored an http logo_uri and this refused to fetch it, a
	// test project's logo would be permanently unavailable.
	//
	// This is the ONLY relaxation here. The content-type allowlist, the SVG
	// exclusion, the byte cap and the sniffed-type agreement check below all
	// still apply, and the address filter still applies through clientFor's
	// underlying SafeDialer.
	allowHTTP := f.OAuthFeatureConfig.GetClientIDMetadataDocument().IsInsecureHTTPAllowed()
	schemeOK := parseErr == nil && (strings.EqualFold(u.Scheme, "https") ||
		(allowHTTP && strings.EqualFold(u.Scheme, "http")))
	if !schemeOK || u.Host == "" {
		return nil, "", ErrLogoInvalidURI
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	// Accept only what will actually be served, so a well-behaved server
	// does not send an oversized or unsupported format we then discard.
	req.Header.Set("Accept", "image/png, image/jpeg, image/gif, image/webp")
	req.Header.Set("User-Agent", "Authgear")

	resp, err := f.clientFor(ctx, u).Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, "", fmt.Errorf("%w: %d", ErrResponseNotOK, resp.StatusCode)
	}

	// Content-Type from the server, normalized: strip parameters and
	// lowercase, then check the allowlist. mime.ParseMediaType handles
	// "image/png; charset=utf-8" and rejects garbage. Checked before the
	// body is read, so a declared-unsupported type is rejected cheaply.
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || !allowedLogoContentTypes[strings.ToLower(mediaType)] {
		return nil, "", ErrLogoContentType
	}

	// Progressive size enforcement, same +1 trick as the document fetch.
	// Content-Length is never trusted.
	body, err = io.ReadAll(io.LimitReader(resp.Body, MaxLogoBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) > MaxLogoBytes {
		return nil, "", ErrLogoTooLarge
	}

	// Sniff the bytes and require agreement with the declared type. A
	// server can lie in Content-Type; the bytes are what a browser will act
	// on. http.DetectContentType reads at most 512 bytes and returns one of
	// a fixed set including the four image types above.
	sniffed, _, err := mime.ParseMediaType(http.DetectContentType(body))
	if err != nil || !allowedLogoContentTypes[strings.ToLower(sniffed)] {
		return nil, "", ErrLogoContentMismatch
	}

	// Serve the SNIFFED type, not the declared one: the header sent to the
	// browser is what the browser trusts, so it must be derived from the
	// bytes, not from a server that could lie.
	return body, sniffed, nil
}

// clientFor mirrors Fetcher.clientFor exactly -- same flag, same two
// clients, same warning log -- kept as a separate method (not shared code)
// so LogoFetcher and Fetcher remain independently readable siblings, per
// the type's own doc comment.
func (f *LogoFetcher) clientFor(ctx context.Context, u *url.URL) *http.Client {
	if !f.OAuthFeatureConfig.GetClientIDMetadataDocument().IsInsecureFetchAddressAllowed() {
		return f.HTTPClients.Strict
	}
	logger := FetcherLogger.GetLogger(ctx)
	logger.Warn(ctx, "cimd: fetching a client logo with SSRF address protection disabled",
		slog.String("app_id", string(f.AppID)),
		slog.String("host", u.Hostname()),
		slog.String("flag", "oauth.client_id_metadata_document.insecure_fetch_address_allowed"),
	)
	return f.HTTPClients.Insecure
}
