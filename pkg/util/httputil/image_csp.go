package httputil

import (
	"net/http"
)

// ImageCSP sandboxes any response that a browser ends up treating as a
// document.
//
// The images endpoint serves arbitrary uploaded bytes and is same origin with
// the Auth UI in the default deployment. Handlers are responsible for never
// declaring an active Content-Type, and this is the backstop if one slips
// through: as a document the response gets a unique origin and no script
// execution, and as an <img> subresource the directives do not apply.
func ImageCSP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
		next.ServeHTTP(w, r)
	})
}
