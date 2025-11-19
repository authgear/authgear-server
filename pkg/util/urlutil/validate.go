package urlutil

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"unicode"
)

var (
	ErrEmptyURL     = errors.New("empty url")
	ErrTooLong      = errors.New("url too long")
	ErrControlChars = errors.New("url contains control characters")
	ErrNotHTTPS     = errors.New("url scheme must be https")
	ErrUserInfo     = errors.New("url must not contain userinfo")
	ErrBadHost      = errors.New("invalid or missing host")
	ErrIPNotAllowed = errors.New("ip addresses are not allowed")
	ErrBlockedHost  = errors.New("host is blocked by policy")
	ErrSingleLabel  = errors.New("single-label hostnames are not allowed (must contain a dot)")
	ErrFragment     = errors.New("url must not contain fragment")
)

const MaxURLLength = 2000 // Maximum allowed URL length

var strictBlocked = []string{
	// Local & special-use
	"localhost",
	"local",
	"localdomain",

	// Internal / enterprise
	"internal",
	"intranet",
	"corp",
	"lan",
	"domain",
	"private",
	"home.arpa",

	// Reserved / test / documentation
	"example",
	"example.com",
	"example.net",
	"example.org",
	"test",
	"invalid",

	// Kubernetes cluster names
	"cluster.local",
	"svc",
	"svc.cluster.local",

	// Cloud metadata hostnames
	"metadata.google.internal",
	"instance-data",

	// Common LAN device names
	"router",
	"gateway",
	"modem",
	"printer",
	"nas",
}

// -----------------------------------------------------------------------------
// ValidateHTTPSStrict
//
// Rules enforced:
//
// 1. Reject empty or whitespace-only URLs
// 2. Reject URLs containing control characters
// 3. Only allow HTTPS scheme
// 4. Reject URLs containing userinfo (user:pass@host)
// 5. Host must exist and be valid
// 6. Reject hosts that are IP addresses (IPv4 or IPv6)
// 7. Reject single-label hostnames (must contain a dot)
// 8. Blocklist: reject blocked hostnames (exact or suffix)
// 9. Validate hostname characters (letters / digits / hyphen / dot)
// 10. Reject URLs containing fragments (#...). Fragments are client-side only, may contain tracking data, and provide no value for server-side URL validation
// 11. Enforce max URL length using constant MaxURLLength
// -----------------------------------------------------------------------------
func ValidateHTTPSStrict(raw string) error {

	// 11. Enforce maximum URL length
	if len(raw) > MaxURLLength {
		return ErrTooLong
	}

	// 1. Trim and reject empty URLs
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ErrEmptyURL
	}

	// 2. Reject control characters
	for _, r := range raw {
		if r == '\r' || r == '\n' || unicode.IsControl(r) {
			return ErrControlChars
		}
	}

	// Parse URL
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}

	// 10. Reject URLs with fragments
	if u.Fragment != "" {
		return ErrFragment
	}

	// 3. Only allow HTTPS
	if strings.ToLower(u.Scheme) != "https" {
		return ErrNotHTTPS
	}

	// 4. Reject userinfo
	if u.User != nil {
		return ErrUserInfo
	}

	// 5. Host must exist
	if u.Host == "" {
		return ErrBadHost
	}

	hostOnly := u.Host
	if h, _, err2 := net.SplitHostPort(u.Host); err2 == nil {
		hostOnly = h
	}
	hostOnly = strings.Trim(hostOnly, "[]")
	if hostOnly == "" {
		return ErrBadHost
	}

	lhost := strings.ToLower(hostOnly)

	// 6. Reject IP addresses
	if ip := net.ParseIP(lhost); ip != nil {
		return ErrIPNotAllowed
	}

	// 7. Reject single-label hostnames
	if !strings.Contains(lhost, ".") {
		return ErrSingleLabel
	}

	// 8. Blocklist checks
	for _, blocked := range strictBlocked {
		b := strings.ToLower(blocked)
		if lhost == b || strings.HasSuffix(lhost, "."+b) {
			return ErrBlockedHost
		}
	}

	// 9. Validate hostname characters
	if !isHostnameSafe(hostOnly) {
		return ErrBadHost
	}

	return nil
}

// ref https://www.ietf.org/rfc/rfc1035.txt
func isHostnameSafe(h string) bool {
	h = strings.TrimSuffix(h, ".")

	if h == "" {
		return false
	}

	for _, r := range h {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.') {
			return false
		}
	}

	labels := strings.Split(h, ".")
	for _, lbl := range labels {
		if lbl == "" {
			return false
		}
		if strings.HasPrefix(lbl, "-") || strings.HasSuffix(lbl, "-") {
			return false
		}
	}

	return true
}
