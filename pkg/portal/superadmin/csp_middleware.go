package superadmin

import "net/http"

func SuperadminCSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; "+
				"script-src 'self' https: 'strict-dynamic'; "+
				"object-src 'none'; "+
				"base-uri 'none'; "+
				"frame-ancestors 'none';")
		next.ServeHTTP(w, r)
	})
}
