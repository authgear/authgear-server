package httputil

import (
	"net/http"
	"strconv"
)

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	body := []byte("OK")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
