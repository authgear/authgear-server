package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigurePlansListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/plans")
}

type PlansListHandler struct {
	// Service dependencies added in Stage 5
}

func (h *PlansListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
