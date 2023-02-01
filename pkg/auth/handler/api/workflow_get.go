package api

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureWorkflowGetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/api/workflow/v1/:instanceid")
}

type WorkflowGetHandler struct{}

func (h *WorkflowGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	instanceID := httproute.GetParam(r, "instanceid")
	w.Write([]byte(fmt.Sprintf("workflow get %s", instanceID)))
}
