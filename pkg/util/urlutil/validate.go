package urlutil

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var (
	ErrTooLong      = errors.New("url too long")
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
// - Reject empty URLs
// - Reject URLs containing control characters
// - Only allow HTTPS scheme
// - Reject URLs containing userinfo (user:pass@host)
// - Host must exist and be valid
// - Reject hosts that are IP addresses (IPv4 or IPv6)
// - Reject single-label hostnames (must contain a dot)
// - Blocklist: reject blocked hostnames (exact or suffix)
// - Validate hostname characters (letters / digits / hyphen / dot)
// - Reject URLs containing fragments (#...). Fragments are client-side only, may contain tracking data, and provide no value for server-side URL validation
// - Enforce max URL length using constant MaxURLLength
// -----------------------------------------------------------------------------
func ValidateHTTPSStrict(raw string) error {

	// Enforce maximum URL length
	if len(raw) > MaxURLLength {
		return ErrTooLong
	}

	// Parse URL
	// Reject URLs containing control characters
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}

	// Reject URLs with fragments
	if u.Fragment != "" {
		return ErrFragment
	}

	// Host must exist
	if u.Host == "" {
		return ErrBadHost
	}

	// Only allow HTTPS
	if strings.ToLower(u.Scheme) != "https" {
		return ErrNotHTTPS
	}

	// Reject userinfo
	if u.User != nil {
		return ErrUserInfo
	}

	hostname := u.Hostname()
	lhost := strings.ToLower(hostname)

	// Reject IP addresses
	if net.ParseIP(hostname) != nil {
		return ErrIPNotAllowed
	}

	// Reject single-label hostnames
	if !strings.Contains(lhost, ".") {
		return ErrSingleLabel
	}

	// Blocklist checks
	for _, blocked := range strictBlocked {
		b := strings.ToLower(blocked)
		if lhost == b || strings.HasSuffix(lhost, "."+b) {
			return ErrBlockedHost
		}
	}

	// Validate hostname characters
	if !isHostnameSafe(hostname) {
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
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.') {
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
