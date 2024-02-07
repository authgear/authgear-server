package httputil

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
)

type GzipMiddleware struct{}

func (m GzipMiddleware) Handle(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(next)
}
