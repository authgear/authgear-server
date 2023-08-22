package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflow2V1GetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v1/workflow2s/get")
}

var Workflow2V1GetRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"id": { "type": "string" }
		},
		"required": ["id"]
	}
`)

type Workflow2V1GetRequest struct {
	ID string `json:"id,omitempty"`
}

type Workflow2V1GetHandler struct {
	JSON           JSONResponseWriter
	Cookies        Workflow2V1CookieManager
	Workflows      Workflow2V1WorkflowService
	OAuthSessions  Workflow2V1OAuthSessionService
	UIInfoResolver Workflow2V1UIInfoResolver
}

func (h *Workflow2V1GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request Workflow2V1GetRequest
	err = httputil.BindJSONBody(r, w, Workflow2V1GetRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	instanceID := request.ID
	userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.Workflows.Get(instanceID, userAgentID)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	result := workflow.FlowResponse{
		ID:     output.Workflow.InstanceID,
		Data:   output.Data,
		Schema: output.SchemaBuilder,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}
