package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
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
	Get(ctx context.Context, workflowID string, instanceID string, userAgentID string) (*workflow.ServiceOutput, error)
	FeedInput(ctx context.Context, workflowID string, instanceID string, userAgentID string, input workflow.Input) (*workflow.ServiceOutput, error)
}

type WorkflowInputCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type WorkflowInputHandler struct {
	JSON      JSONResponseWriter
	Workflows WorkflowInputWorkflowService
	Cookies   WorkflowNewCookieManager
}

func (h *WorkflowInputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workflowID := httproute.GetParam(r, "workflowid")
	instanceID := httproute.GetParam(r, "instanceid")

	var err error
	var request WorkflowInputRequest

	err = httputil.BindJSONBody(r, w, WorkflowInputRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.handle(ctx, w, r, workflowID, instanceID, userAgentID, request)
	if err != nil {
		apiResp, apiRespErr := h.prepareErrorResponse(ctx, workflowID, instanceID, userAgentID, err)
		if apiRespErr != nil {
			// failed to get the workflow when preparing the error response
			h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
			return
		}
		h.JSON.WriteResponse(w, apiResp)
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
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request WorkflowInputRequest,
) (*workflow.ServiceOutput, error) {
	input, err := workflow.InstantiateInputFromPublicRegistry(request.Input)
	if err != nil {
		return nil, err
	}

	output, err := h.Workflows.FeedInput(ctx, workflowID, instanceID, userAgentID, input)
	if err != nil && errors.Is(err, workflow.ErrNoChange) {
		err = workflow.ErrInvalidInputKind
	}
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *WorkflowInputHandler) prepareErrorResponse(
	ctx context.Context,
	workflowID string,
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(ctx, workflowID, instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}
