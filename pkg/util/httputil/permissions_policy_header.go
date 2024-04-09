package httputil

import (
	"net/http"
)

func PermissionsPolicyHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpPermissionsPolicy := HTTPPermissionsPolicy(DefaultPermissionsPolicy)
		w.Header().Set("Permissions-Policy", httpPermissionsPolicy.String())
		next.ServeHTTP(w, r)
	})
}
