package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func deriveFrameAncestors(clients []config.OAuthClientConfiguration) (out []string) {
	for _, client := range clients {
		if redirectURIs, ok := client["redirect_uris"].([]interface{}); ok {
			for _, redirectURI := range redirectURIs {
				if s, ok := redirectURI.(string); ok {
					u, err := url.Parse(s)
					if err == nil {
						if u.Host != "" && (u.Scheme == "http" || u.Scheme == "https") {
							ancestor := url.URL{
								Scheme: u.Scheme,
								Host:   u.Host,
							}
							out = append(out, ancestor.String())
						}
					}
				}
			}
		}
	}
	return
}

// CSPMiddleware derives frame-ancestors from clients and
// writes Content-Security-Policy.
type CSPMiddleware struct {
	Clients []config.OAuthClientConfiguration
}

func (m *CSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		frameAncestors := deriveFrameAncestors(m.Clients)
		frameAncestors = append(frameAncestors, "'self'")
		csp := fmt.Sprintf("frame-ancestors %s;", strings.Join(frameAncestors, " "))
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}
