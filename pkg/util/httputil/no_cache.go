package httputil

import (
	"net/http"
)

// NoCache allows caches to store a response but requires them to
// revalidate it before reuse.
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Pragma", "no-cache")
		next.ServeHTTP(w, r)
	})
}
