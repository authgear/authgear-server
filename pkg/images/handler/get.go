package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/imageproxy"
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

type GetHandler struct {
	Director imageproxy.Director
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Director == nil {
		http.Error(w, "images are disabled", http.StatusInternalServerError)
		return
	}

	director := h.Director.Director
	reverseProxy := httputil.ReverseProxy{
		Director: director,
		// ErrorHandler
		// ModifyResponse
	}
	reverseProxy.ServeHTTP(w, r)
}
