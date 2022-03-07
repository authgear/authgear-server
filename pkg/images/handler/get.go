package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/imageproxy"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureGetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/_images/:appid/:objectid/:options")
}

var ExtractKey imageproxy.ExtractKey = func(r *http.Request) string {
	return fmt.Sprintf(
		"%s/%s",
		httproute.GetParam(r, "appid"),
		httproute.GetParam(r, "objectid"),
	)
}

type GetHandlerLogger struct{ *log.Logger }

func NewGetHandlerLogger(lf *log.Factory) GetHandlerLogger {
	return GetHandlerLogger{lf.New("get-handler")}
}

type GetHandler struct {
	Director imageproxy.Director
	Logger   GetHandlerLogger
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Director == nil {
		http.Error(w, "images are disabled", http.StatusInternalServerError)
		return
	}

	director := h.Director.Director
	reverseProxy := httputil.ReverseProxy{
		Director: director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			h.Logger.WithError(err).Errorf("reverse proxy error")
			w.WriteHeader(http.StatusBadGateway)
		},
		// ModifyResponse
	}
	reverseProxy.ServeHTTP(w, r)
}
