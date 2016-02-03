package router

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func CorsMiddleware(corsOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsMethod := r.Header.Get("Access-Control-Request-Method")
		corsHeaders := r.Header.Get("Access-Control-Request-Headers")

		log.Debugf("CORS Method: %s", corsMethod)
		log.Debugf("CORS Headers: %s", corsHeaders)

		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", corsMethod)
		w.Header().Set("Access-Control-Allow-Headers", corsHeaders)

		next.ServeHTTP(w, r)
	})
}
