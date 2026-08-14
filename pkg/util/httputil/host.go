package httputil

import (
	"net/http"
	"strings"
)

type HTTPHost string

func GetHost(r *http.Request, trustProxy bool) string {
	if trustProxy {
		// X-Forwarded-Host may contain a comma-separated list of hosts
		// when the request passes through a chain of proxies, each
		// appending its own value. The first entry is the one closest
		// to the original client, so take that one.
		if host := r.Header.Get("X-Forwarded-Host"); host != "" {
			parts := strings.Split(host, ",")
			return strings.TrimSpace(parts[0])
		}

		if host := r.Header.Get("X-Original-Host"); host != "" {
			parts := strings.Split(host, ",")
			return strings.TrimSpace(parts[0])
		}
	}

	return r.Host
}
