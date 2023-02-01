package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureWorkflowNewRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/workflow/v1")
}

type WorkflowNewHandler struct{}

func (h *WorkflowNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("workflow new"))
}
