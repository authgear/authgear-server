package httputil

import (
	"net/http"

	"github.com/go-http-utils/etag"
)

func ETag(next http.Handler) http.Handler {
	return etag.Handler(next, true /* weak */)
}
