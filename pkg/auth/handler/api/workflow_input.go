package api

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflowInputRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/workflow/v1/:instanceid")
}

var WorkflowInputRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"input": {
				"type": "object",
				"properties": {
					"kind": { "type": "string" },
					"data": { "type": "object" }
				},
				"required": ["kind", "data"]
			}
		},
		"required": ["input"]
	}
`)

type WorkflowInputRequest struct {
	Input workflow.InputJSON `json:"input"`
}

type WorkflowInputWorkflowService interface {
	FeedInput(instanceID string, input workflow.Input) (*workflow.ServiceOutput, error)
}

type WorkflowInputHandler struct {
	Database  *appdb.Handle
	JSON      JSONResponseWriter
	Workflows WorkflowInputWorkflowService
}

func (h *WorkflowInputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	instanceID := httproute.GetParam(r, "instanceid")

	var err error
	var request WorkflowInputRequest

	err = httputil.BindJSONBody(r, w, WorkflowInputRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	var output *workflow.ServiceOutput
	err = h.Database.WithTx(func() error {
		output, err = h.handle(w, r, instanceID, request)
		return err
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

func (h *WorkflowInputHandler) handle(w http.ResponseWriter, r *http.Request, instanceID string, request WorkflowInputRequest) (*workflow.ServiceOutput, error) {
	input, err := workflow.InstantiateInput(request.Input)
	if err != nil {
		return nil, err
	}

	output, err := h.Workflows.FeedInput(instanceID, input)
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}
