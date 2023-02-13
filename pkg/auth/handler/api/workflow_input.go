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
		WithPathPattern("/api/v1/workflows/:workflowid/instances/:instanceid")
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
	FeedInput(workflowID string, instanceID string, input workflow.Input) (*workflow.ServiceOutput, error)
}

type WorkflowInputHandler struct {
	Database  *appdb.Handle
	JSON      JSONResponseWriter
	Workflows WorkflowInputWorkflowService
}

func (h *WorkflowInputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	workflowID := httproute.GetParam(r, "workflowid")
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
		output, err = h.handle(w, r, workflowID, instanceID, request)
		return err
	})
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *WorkflowInputHandler) handle(
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	request WorkflowInputRequest,
) (*workflow.ServiceOutput, error) {
	input, err := workflow.InstantiateInputFromPublicRegistry(request.Input)
	if err != nil {
		return nil, err
	}

	output, err := h.Workflows.FeedInput(workflowID, instanceID, input)
	if err != nil && errors.Is(err, workflow.ErrNoChange) {
		err = workflow.ErrInvalidInputKind
	}
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}
