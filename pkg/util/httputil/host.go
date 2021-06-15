package httputil

import (
	"net/http"
)

func GetHost(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if host := r.Header.Get("X-Forwarded-Host"); host != "" {
			return host
		}

		if host := r.Header.Get("X-Original-Host"); host != "" {
			return host
		}
	}

	return r.Host
}
