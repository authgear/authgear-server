package oauthclient

import (
	"errors"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

const (
	DCRClientIDPrefix = "dcrc_"

	dcrClientIDAlphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	dcrClientIDRandomLength = 22 // matches spec: "22 chars URL-safe base64 (16 bytes)"
)

// GenerateDCRClientID returns a new dcrc_-prefixed client_id for a
// DCR-registered client.
func GenerateDCRClientID() string {
	return DCRClientIDPrefix + rand.StringWithAlphabet(dcrClientIDRandomLength, dcrClientIDAlphabet, rand.SecureRand)
}

// IsDCRClientID reports whether clientID has the shape of a DCR-registered
// client_id, as opposed to a statically configured or CIMD client_id.
func IsDCRClientID(clientID string) bool {
	return strings.HasPrefix(clientID, DCRClientIDPrefix)
}

var (
	ErrCIMDClientIDNotHTTPS      = errors.New("oauthclient: cimd client_id must use the https scheme")
	ErrCIMDClientIDNoPath        = errors.New("oauthclient: cimd client_id must contain a path component")
	ErrCIMDClientIDHasFragment   = errors.New("oauthclient: cimd client_id must not contain a fragment")
	ErrCIMDClientIDHasUserInfo   = errors.New("oauthclient: cimd client_id must not contain userinfo")
	ErrCIMDClientIDHasDotSegment = errors.New("oauthclient: cimd client_id must not contain '.' or '..' path segments")
	ErrCIMDClientIDBadHost       = errors.New("oauthclient: cimd client_id has an invalid or missing host")
	ErrCIMDClientIDTooLong       = errors.New("oauthclient: cimd client_id is too long")
)

const MaxCIMDClientIDLength = urlutil.MaxURLLength

// ParseCIMDClientID implements docs/specs/cimd.md § Client ID Format. It
// performs no network access: a malformed client_id must never reach the
// fetch path. The returned URL is what the fetcher requests, unmodified --
// no normalization, because the document's own client_id must equal this
// string byte-for-byte.
//
// allowInsecureHTTP is an explicit parameter rather than a config read
// inside the function, so the function stays pure and every call site is
// compiler-forced to state its posture. It widens only the scheme rule --
// every other check still applies.
func ParseCIMDClientID(clientID string, allowInsecureHTTP bool) (*url.URL, error) {
	if len(clientID) > MaxCIMDClientIDLength {
		return nil, ErrCIMDClientIDTooLong
	}

	u, err := url.Parse(clientID)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.EqualFold(u.Scheme, "https"):
	case allowInsecureHTTP && strings.EqualFold(u.Scheme, "http"):
	default:
		return nil, ErrCIMDClientIDNotHTTPS
	}
	if u.User != nil {
		return nil, ErrCIMDClientIDHasUserInfo
	}
	// strings.Contains on the raw string, not u.Fragment/u.RawFragment: Go's
	// net/url drops a bare trailing "#" entirely (no ForceQuery-style flag
	// for it), so a URL.Fragment-based check would silently accept one.
	if strings.Contains(clientID, "#") {
		return nil, ErrCIMDClientIDHasFragment
	}
	if u.Host == "" || u.Hostname() == "" {
		return nil, ErrCIMDClientIDBadHost
	}
	if u.Path == "" {
		return nil, ErrCIMDClientIDNoPath
	}
	// u.Path is the DECODED path, so this also catches "%2e%2e".
	for _, seg := range strings.Split(u.Path, "/") {
		if seg == "." || seg == ".." {
			return nil, ErrCIMDClientIDHasDotSegment
		}
	}

	return u, nil
}

// IsCIMDClientID reports whether clientID has the shape of a valid CIMD
// Client Identifier URL. It applies no trust policy (allowed_domains) and
// no address-reachability check -- see IsCIMDClientIDAllowed for the
// former and the fetcher's address filter for the latter.
func IsCIMDClientID(clientID string, allowInsecureHTTP bool) bool {
	_, err := ParseCIMDClientID(clientID, allowInsecureHTTP)
	return err == nil
}

// IsCIMDClientIDAllowed reports whether u's host satisfies the project's
// allowed_domains policy (docs/specs/cimd.md § Domain Trust). An empty
// policy allows everything. Matching is on Hostname() only -- never
// host:port -- and is case-insensitive.
//
// A leading "*." matches exactly ONE label, per the RFC 6125 / TLS
// certificate convention: "*.example.com" matches "a.example.com" but not
// "a.b.example.com" and not "example.com".
func IsCIMDClientIDAllowed(allowedDomains []string, u *url.URL) bool {
	if len(allowedDomains) == 0 {
		return true
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	for _, pattern := range allowedDomains {
		p := strings.ToLower(pattern)
		if suffix, ok := strings.CutPrefix(p, "*."); ok {
			rest, found := strings.CutSuffix(host, "."+suffix)
			if found && rest != "" && !strings.Contains(rest, ".") {
				return true
			}
			continue
		}
		if host == p {
			return true
		}
	}
	return false
}
