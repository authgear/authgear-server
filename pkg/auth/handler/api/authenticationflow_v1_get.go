package api

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAuthenticationFlowV1GetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/authentication_flows/states")
}

type AuthenticationFlowV1NonRestfulGetRequest struct {
	StateToken string `json:"state_token,omitempty"`
}

var AuthenticationFlowV1NonRestfulGetRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"state_token": { "type": "string" }
		},
		"required": ["state_token"]
	}
`)

type AuthenticationFlowV1GetHandler struct {
	RedisHandle *appredis.Handle
	Workflows   AuthenticationFlowV1WorkflowService
}

func (h *AuthenticationFlowV1GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AuthenticationFlowV1NonRestfulGetRequest
	ctx := r.Context()
	err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulGetRequestSchema.Validator(), &request)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	stateToken := request.StateToken
	h.get(ctx, w, r, stateToken)
}

func (h *AuthenticationFlowV1GetHandler) get(ctx context.Context, w http.ResponseWriter, r *http.Request, stateToken string) {
	output, err := h.Workflows.Get(ctx, stateToken)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	result := output.ToFlowResponse()
	httputil.WriteJSONResponse(ctx, w, &api.Response{Result: result})
}
