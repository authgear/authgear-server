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

				// In case the downstream middlewares do not produce a response,
				// We write a HTTP 500 here.
				// If the response has been written, this has effect but will trigger a warning.
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
