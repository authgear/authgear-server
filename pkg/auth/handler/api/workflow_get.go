package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureWorkflowGetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/api/workflow/v1/:instanceid")
}

type WorkflowGetWorkflowService interface {
	Get(instanceID string) (*workflow.ServiceOutput, error)
}

type WorkflowGetHandler struct {
	Database  *appdb.Handle
	JSON      JSONResponseWriter
	Workflows WorkflowGetWorkflowService
}

func (h *WorkflowGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	instanceID := httproute.GetParam(r, "instanceid")

	var output *workflow.ServiceOutput
	var err error
	err = h.Database.WithTx(func() error {
		output, err = h.Workflows.Get(instanceID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}
