package webapp

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/config"
)

func deriveFrameAncestors(clients []config.OAuthClientConfig) (out []string) {
	for _, client := range clients {
		if redirectURIs, ok := client["redirect_uris"].([]interface{}); ok {
			for _, redirectURI := range redirectURIs {
				if s, ok := redirectURI.(string); ok {
					u, err := url.Parse(s)
					if err == nil {
						ancestor := (&url.URL{
							Scheme: u.Scheme,
							Host:   u.Host,
						}).String()
						if u.Scheme == "https" && u.Host != "" {
							out = append(out, ancestor)
						}
						if u.Scheme == "http" && isLocalhost(u.Host) {
							out = append(out, ancestor)
						}
					}
				}
			}
		}
	}
	return
}

func isLocalhost(host string) bool {
	// Trim the port if it is present.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// net.ParseIP does not accept IPv6 in square brackets.
	// Remove them first.
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	ip := net.ParseIP(host)
	// host is either IPv4 or IPv6
	if ip != nil {
		return ip.IsLoopback()
	}

	// host is a hostname.
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	return false
}

// CSPMiddleware derives frame-ancestors from clients and
// writes Content-Security-Policy.
type CSPMiddleware struct {
	Config *config.OAuthConfig
}

func (m *CSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		frameAncestors := deriveFrameAncestors(m.Config.Clients)
		frameAncestors = append(frameAncestors, "'self'")
		csp := fmt.Sprintf("frame-ancestors %s;", strings.Join(frameAncestors, " "))
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}
