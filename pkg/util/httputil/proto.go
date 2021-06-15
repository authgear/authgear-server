package httputil

import (
	"net/http"
)

func GetProto(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			return proto
		}

		if proto := r.Header.Get("X-Original-Proto"); proto != "" {
			return proto
		}
	}

	if r.TLS != nil {
		return "https"
	}

	return "http"
}
