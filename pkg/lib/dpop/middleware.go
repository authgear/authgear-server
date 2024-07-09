package dpop

import "net/http"

type Middleware struct{}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// TODO
		next.ServeHTTP(rw, r)
	})
}
