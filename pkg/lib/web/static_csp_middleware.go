package web

import (
	"net/http"
)

type StaticCSPMiddleware struct {
	CSPDirectives []string
}

func (m StaticCSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", CSPJoin(m.CSPDirectives))
		next.ServeHTTP(w, r)
	})
}
