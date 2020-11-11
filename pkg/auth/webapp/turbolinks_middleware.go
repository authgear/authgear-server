package webapp

import (
	"net/http"
)

type TurbolinksMiddleware struct{}

func (m TurbolinksMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Turbolinks-Location", r.URL.String())
		next.ServeHTTP(w, r)
	})
}
