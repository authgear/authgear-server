package httputil

// IsLoopbackHost reports whether host -- as returned by url.URL.Hostname(),
// which strips the brackets from an IPv6 literal like "[::1]:3000" -- names
// a loopback interface: "localhost", or the IPv4/IPv6 loopback literals
// "127.0.0.1" and "::1". Per RFC 8252 §7.3, native/CLI OAuth clients use any
// of these interchangeably for their redirect_uri callback listener.
//
// This checks the literal hostname string, not a resolved IP address: it is
// for validating a redirect_uri that Authgear itself never connects to
// (unlike the SSRF concern IsPubliclyRoutable addresses for URLs Authgear
// fetches), so there is no DNS resolution step to filter.
func IsLoopbackHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
