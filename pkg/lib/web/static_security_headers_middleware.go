package web

import "net/http"

type StaticSecurityHeadersMiddleware struct{}

func (StaticSecurityHeadersMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
