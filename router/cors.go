package router

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type CORSMiddleware struct {
	Origin string
	Next   http.Handler
}

func (cors *CORSMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestMethod := r.Method
	corsMethod := r.Header.Get("Access-Control-Request-Method")
	corsHeaders := r.Header.Get("Access-Control-Request-Headers")

	w.Header().Set("Access-Control-Allow-Origin", cors.Origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if corsMethod != "" {
		log.Debugf("CORS Method: %s", corsMethod)
		w.Header().Set("Access-Control-Allow-Methods", corsMethod)
	}

	if corsHeaders != "" {
		log.Debugf("CORS Headers: %s", corsHeaders)
		w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
	}

	if requestMethod == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
	} else {
		cors.Next.ServeHTTP(w, r)
	}
}
