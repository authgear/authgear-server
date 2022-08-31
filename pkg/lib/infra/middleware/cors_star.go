package middleware

import (
	"net/http"
)

func CORSStar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "900") // 15 mins

		corsMethod := r.Header.Get("Access-Control-Request-Method")
		if corsMethod != "" {
			w.Header().Set("Access-Control-Allow-Methods", corsMethod)
		}

		corsHeaders := r.Header.Get("Access-Control-Request-Headers")
		if corsHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
		}

		requestMethod := r.Method
		if requestMethod == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
