package middleware

import (
	"net/http"
)

// MaxBodySize is the maximum size of HTTP request body: 1MB.
const MaxBodySize = 1 //* 1024 * 1024

type BodyLimitMiddleware struct {
}

func (m *BodyLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
		next.ServeHTTP(w, r)
	})
}
