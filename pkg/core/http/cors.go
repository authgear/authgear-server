package http

import (
	"net/http"
	"strings"
)

func parseVaryHeaders(values []string) []string {
	var headers []string
	for _, v := range values {
		for _, h := range strings.Split(v, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				headers = append(headers, h)
			}
		}
	}
	return headers
}

func FixupCORSHeaders(downstream http.ResponseWriter, upstream *http.Response) {
	hasCORSHeaders := false
	for name, values := range upstream.Header {
		if strings.HasPrefix(name, "Access-Control-") {
			hasCORSHeaders = true
			break
		}
		if name == "Vary" {
			varyHeaders := parseVaryHeaders(values)
			for _, h := range varyHeaders {
				if http.CanonicalHeaderKey(h) == "Origin" {
					hasCORSHeaders = true
				}
			}
		}
	}

	if !hasCORSHeaders {
		return
	}

	// Upstream has provided CORS header; upstream will manage all CORS headers
	// Remove existing CORS headers from response to downstream
	// `Vary: Origin` is also used to control CORS policy, so upstream
	// should manage it themselves.
	headers := downstream.Header()
	for name, values := range headers {
		if strings.HasPrefix(name, "Access-Control-") {
			headers.Del(name)
		}
		// Delete 'Vary: Origin' header
		if name == "Vary" {
			varyHeaders := parseVaryHeaders(values)
			n := 0
			for _, h := range varyHeaders {
				if http.CanonicalHeaderKey(h) != "Origin" {
					varyHeaders[n] = h
					n++
				}
			}
			varyHeaders = varyHeaders[:n]
			if len(varyHeaders) > 0 {
				headers[name] = []string{strings.Join(varyHeaders, ",")}
			} else {
				delete(headers, name)
			}
		}
	}
}
