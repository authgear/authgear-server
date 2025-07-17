package middleware

import (
	"net/http"
)

// CORSMiddleware provides CORS headers by matching request origin with the configured allowed origins
// The allowed origins are provided through app config and environment variable
type CORSMiddleware struct {
	Matcher *CORSMatcher
}

func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matcher, err := m.Matcher.PrepareOriginMatcher(r)
		// nolint: staticcheck
		if err != nil {
			// err is handled by not writing any CORS headers.
		}

		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")
		if origin != "" && err == nil && matcher.MatchOrigin(origin) {
			corsMethod := r.Header.Get("Access-Control-Request-Method")
			corsHeaders := r.Header.Get("Access-Control-Request-Headers")

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "900") // 15 mins

			if corsMethod != "" {
				w.Header().Set("Access-Control-Allow-Methods", corsMethod)
			}

			if corsHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
			}
		}

		requestMethod := r.Method
		if requestMethod == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
