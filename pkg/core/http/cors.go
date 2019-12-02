package http

import (
	"net/http"
	"strings"
)

func FixupCORSHeaders(downstream http.ResponseWriter, upstream *http.Response) {
	hasCORSHeaders := false
	for name := range upstream.Header {
		if strings.HasPrefix(name, "Access-Control-") {
			hasCORSHeaders = true
			break
		}
	}

	if !hasCORSHeaders {
		return
	}

	// Upstream has provided CORS header; upstream will manage all CORS headers
	// Remove existing CORS headers from response to downstream
	headers := downstream.Header()
	for name := range headers {
		if strings.HasPrefix(name, "Access-Control-") {
			headers.Del(name)
		}
	}
}
