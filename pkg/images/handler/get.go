package handler

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/_images/:appid/:objectid/:options")
}

type GetHandler struct {
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO(images): get image endpint
	http.Error(w, "not implemented", http.StatusBadRequest)
}
