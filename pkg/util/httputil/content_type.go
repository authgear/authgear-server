package httputil

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func CheckContentType(raws []string) httproute.MiddlewareFunc {
	allowedMediaTypes := getAllowedMediaType(raws)

	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return makeCheckContentTypeHandlerFunc(next, allowedMediaTypes)
	})
}

func getAllowedMediaType(raws []string) []string {
	var allowedMediaTypes []string
	for _, raw := range raws {
		mediaType, _, err := mime.ParseMediaType(raw)
		if err != nil {
			panic(fmt.Errorf("httputil: invalid content type: %w", err))
		}
		allowedMediaTypes = append(allowedMediaTypes, mediaType)
	}
	return allowedMediaTypes
}

func makeCheckContentTypeHandlerFunc(next http.Handler, allowedMediaTypes []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For some reason, body is not nil when the request is GET or HEAD.
		if r.Method == "GET" || r.Method == "HEAD" || r.Body == nil {
			// Content-Type is irrelevant without body.
			next.ServeHTTP(w, r)
			return
		}

		raw := r.Header.Get("Content-Type")

		requestContentType, params, err := mime.ParseMediaType(raw)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid content type: %v", raw), http.StatusUnsupportedMediaType)
			return
		}

		isAllowed := false
		for _, allowedMediaType := range allowedMediaTypes {
			if allowedMediaType == requestContentType {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			http.Error(w, fmt.Sprintf("invalid content type: %v", requestContentType), http.StatusUnsupportedMediaType)
			return
		}

		// In case charset is specified, ensure it is utf-8.
		if charset, ok := params["charset"]; ok {
			charset = strings.ToLower(charset)
			if charset != "utf-8" && charset != "utf8" {
				http.Error(w, fmt.Sprintf("invalid content type: %v", requestContentType), http.StatusUnsupportedMediaType)
				return
			}
		}
		// Allow params to contain something we do not understand.

		next.ServeHTTP(w, r)
	})
}
