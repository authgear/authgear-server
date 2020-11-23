package web

import (
	"fmt"
	"net/http"
	"strings"
)

var DefaultStrictCSPDirectives = []string{
	"default-src 'self'",
	"object-src 'none'",
	"base-uri 'none'",
	"block-all-mixed-content",
}

// SecHeadersMiddleware sends default security headers according to configuration.
// TODO(oauth): to support silent token refresh, drive appropriate frame
//              ancestors from OAuth clients.
type SecHeadersMiddleware struct {
	FrameAncestors []string
	CSPDirectives  []string
}

func (m *SecHeadersMiddleware) Handle(next http.Handler) http.Handler {
	header := map[string]string{}
	cspDirectives := make([]string, len(m.CSPDirectives))
	copy(cspDirectives, m.CSPDirectives)

	if len(m.FrameAncestors) == 0 {
		cspDirectives = append(cspDirectives, "frame-ancestors 'none'")
		header["X-Frame-Options"] = "DENY"
	} else {
		fa := fmt.Sprintf("frame-ancestors %s", strings.Join(m.FrameAncestors, " "))
		cspDirectives = append(cspDirectives, fa)
	}

	header["Content-Security-Policy"] = strings.Join(cspDirectives, "; ")
	header["X-Content-Type-Options"] = "nosniff"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range header {
			w.Header().Set(k, v)
		}
		next.ServeHTTP(w, r)
	})
}
