package httputil

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func CheckContentType(raws []string) httproute.MiddlewareFunc {
	var allowedMediaTypes []string
	for _, raw := range raws {
		mediaType, _, err := mime.ParseMediaType(raw)
		if err != nil {
			panic(fmt.Errorf("httputil: invalid content type: %w", err))
		}
		allowedMediaTypes = append(allowedMediaTypes, mediaType)
	}

	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				// Content-Type is irrelevant without body.
				next.ServeHTTP(w, r)
				return
			}

			requestContentType := r.Header.Get("Content-Type")
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

			next.ServeHTTP(w, r)
		})
	})
}
