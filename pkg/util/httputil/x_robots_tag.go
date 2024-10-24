package httputil

import (
	"net/http"
)

func XRobotsTag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "none")
		next.ServeHTTP(w, r)
	})
}
