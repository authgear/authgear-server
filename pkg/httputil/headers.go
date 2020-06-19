package httputil

import "net/http"

func GetHost(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if host := r.Header.Get("X-Forwarded-Host"); host != "" {
			return host
		}
	}

	return r.Host
}

func GetProto(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			return proto
		}
	}

	if r.TLS != nil {
		return "https"
	}

	return "http"
}
