package api

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflowNewRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/workflow/v1")
}

var WorkflowNewRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"intent": {
				"type": "object",
				"properties": {
					"kind": { "type": "string" },
					"data": { "type": "object" }
				},
				"required": ["kind", "data"]
			}
		},
		"required": ["intent"]
	}
`)

type WorkflowNewRequest struct {
	Intent WorkflowNewRequestIntent
}

type WorkflowNewRequestIntent struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type WorkflowNewWorkflowService interface {
	CreateNewWorkflow(intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
}

type WorkflowNewHandler struct {
	Database  *appdb.Handle
	JSON      JSONResponseWriter
	Workflows WorkflowNewWorkflowService
}

func (h *WorkflowNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request WorkflowNewRequest
	err = httputil.BindJSONBody(r, w, WorkflowNewRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	var output *workflow.ServiceOutput
	err = h.Database.WithTx(func() error {
		intent, err := workflow.InstantiateIntent(request.Intent.Kind)
		if err != nil {
			return err
		}

		err = json.Unmarshal(request.Intent.Data, intent)
		if err != nil {
			return err
		}

		// TODO(workflow): sessionOptions
		output, err = h.Workflows.CreateNewWorkflow(intent, &workflow.SessionOptions{
			ClientID: "",
		})
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
		Action: WorkflowAction{
			Type: WorkflowActionTypeContinue,
		},
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}
