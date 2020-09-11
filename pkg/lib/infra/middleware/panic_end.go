package middleware

import (
	"net/http"
)

type PanicEndMiddleware struct{}

func (m *PanicEndMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				_ = err
				// Do NOT rethrow to consume the panic
				// It is assumed that downstream middlewares have handled it somehow.
			}
		}()

		next.ServeHTTP(w, r)
	})
}
