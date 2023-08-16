package api

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflow2V1InputRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v1/workflow2s/input")
}

var Workflow2V1InputRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"required": ["workflow_id", "instance_id"],
		"properties": {
			"workflow_id": { "type": "string" },
			"instance_id": { "type": "string" }
		},
		"oneOf": [
			{
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
			},
			{
				"properties": {
					"batch_input": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"kind": { "type": "string" },
								"data": { "type": "object" }
							},
							"required": ["kind", "data"]
						},
						"minItems": 1
					}
				},
				"required": ["batch_input"]
			}
		]
	}
`)

type Workflow2V1InputRequest struct {
	WorkflowID string                `json:"workflow_id,omitempty"`
	InstanceID string                `json:"instance_id,omitempty"`
	Input      *workflow.InputJSON   `json:"input,omitempty"`
	BatchInput []*workflow.InputJSON `json:"batch_input,omitempty"`
}

type Workflow2V1InputHandler struct {
	JSON           JSONResponseWriter
	Cookies        Workflow2V1CookieManager
	Workflows      Workflow2V1WorkflowService
	OAuthSessions  Workflow2V1OAuthSessionService
	UIInfoResolver Workflow2V1UIInfoResolver
}

func (h *Workflow2V1InputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request Workflow2V1InputRequest
	err = httputil.BindJSONBody(r, w, Workflow2V1InputRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	if request.Input != nil {
		workflowID := request.WorkflowID
		instanceID := request.InstanceID
		userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

		output, err := h.input(w, r, workflowID, instanceID, userAgentID, request)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(workflowID, instanceID, userAgentID, err)
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

		result := Workflow2Response{
			Action:     output.Action,
			WorkflowID: output.Workflow.WorkflowID,
			InstanceID: output.Workflow.InstanceID,
			Data:       output.Data,
		}
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	} else {
		workflowID := request.WorkflowID
		instanceID := request.InstanceID
		userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

		output, err := h.batchInput(w, r, workflowID, instanceID, userAgentID, request)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(workflowID, instanceID, userAgentID, err)
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

		result := Workflow2Response{
			Action:     output.Action,
			WorkflowID: output.Workflow.WorkflowID,
			InstanceID: output.Workflow.InstanceID,
			Data:       output.Data,
		}
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	}
}

func (h *Workflow2V1InputHandler) input(
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request Workflow2V1InputRequest,
) (*workflow.ServiceOutput, error) {
	input, err := workflow.InstantiateInputFromPublicRegistry(*request.Input)
	if err != nil {
		return nil, err
	}

	output, err := h.Workflows.FeedInput(workflowID, instanceID, userAgentID, input)
	if err != nil && errors.Is(err, workflow.ErrNoChange) {
		err = workflow.ErrInvalidInputKind
	}
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *Workflow2V1InputHandler) batchInput(
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request Workflow2V1InputRequest,
) (output *workflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	var input workflow.Input
	for _, inputJSON := range request.BatchInput {
		input, err = workflow.InstantiateInputFromPublicRegistry(*inputJSON)
		if err != nil {
			return nil, err
		}

		output, err = h.Workflows.FeedInput(workflowID, instanceID, userAgentID, input)
		if err != nil && errors.Is(err, workflow.ErrNoChange) {
			err = workflow.ErrInvalidInputKind
		}
		if err != nil && !errors.Is(err, workflow.ErrEOF) {
			return nil, err
		}

		// Feed the next input to the latest instance.
		instanceID = output.Workflow.InstanceID
		cookies = append(cookies, output.Cookies...)
	}
	if err != nil && errors.Is(err, workflow.ErrEOF) {
		err = nil
	}
	if err != nil {
		return
	}

	// Return all collected cookies.
	output.Cookies = cookies
	return
}

func (h *Workflow2V1InputHandler) prepareErrorResponse(
	workflowID string,
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(workflowID, instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := Workflow2Response{
		Action:     output.Action,
		WorkflowID: output.Workflow.WorkflowID,
		InstanceID: output.Workflow.InstanceID,
		Data:       output.Data,
	}
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}
