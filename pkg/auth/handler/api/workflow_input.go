package api

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureWorkflowInputRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/workflow/v1/:instanceid")
}

type WorkflowInputHandler struct{}

func (h *WorkflowInputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	instanceID := httproute.GetParam(r, "instanceid")
	w.Write([]byte(fmt.Sprintf("workflow input %s", instanceID)))
}
