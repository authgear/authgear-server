package httputil

import "strings"

// MatchHostPattern reports whether host matches any of patterns.
//
// Matching is on a hostname only -- never host:port -- and is
// case-insensitive. A leading "*." matches exactly ONE label, per the RFC 6125
// / TLS certificate convention: "*.example.com" matches "a.example.com" but not
// "a.b.example.com" and not the apex "example.com". A single-label hostname
// such as "localhost" is a valid pattern.
//
// An empty pattern list matches nothing. Callers whose empty case means
// "no restriction" check that themselves, so that the two readings of an empty
// list stay at the call site rather than being buried here.
func MatchHostPattern(patterns []string, host string) bool {
	if len(patterns) == 0 {
		return false
	}
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	for _, pattern := range patterns {
		p := strings.ToLower(pattern)
		if suffix, ok := strings.CutPrefix(p, "*."); ok {
			rest, found := strings.CutSuffix(h, "."+suffix)
			if found && rest != "" && !strings.Contains(rest, ".") {
				return true
			}
			continue
		}
		if h == p {
			return true
		}
	}
	return false
}
