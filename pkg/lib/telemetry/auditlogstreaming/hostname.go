package auditlogstreaming

import "net/url"

// ResolveHostname returns the RFC 5424 HOSTNAME for a project: the host of
// its http.public_origin, without the port. HOSTNAME identifies the
// project the entries belong to, not the Authgear process that delivers
// them -- delivery happens in the background worker, whose own hostname
// would tell a receiver nothing.
func ResolveHostname(publicOrigin string) string {
	u, err := url.Parse(publicOrigin)
	if err != nil {
		return "-"
	}
	return sanitizeHostname(u.Hostname())
}

// sanitizeHostname returns "-" unless host is 1-255 characters of printable
// US-ASCII (0x21-0x7E), which is what RFC 5424 permits for HOSTNAME. A
// project whose public_origin uses a non-ASCII internationalised domain
// therefore reports "-"; the project is still identifiable from the app_id
// structured-data parameter. Converting to punycode is deliberately not
// attempted.
func sanitizeHostname(host string) string {
	if len(host) < 1 || len(host) > 255 {
		return "-"
	}
	for i := 0; i < len(host); i++ {
		c := host[i]
		if c < 0x21 || c > 0x7E {
			return "-"
		}
	}
	return host
}
