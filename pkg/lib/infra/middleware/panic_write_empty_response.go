package middleware

import (
	"net/http"
)

type PanicWriteEmptyResponseMiddleware struct{}

func (m *PanicWriteEmptyResponseMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				// Rethrow
				panic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
