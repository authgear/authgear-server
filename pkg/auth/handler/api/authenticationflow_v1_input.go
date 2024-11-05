package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAuthenticationFlowV1InputRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/authentication_flows/states/input")
}

var AuthenticationFlowV1NonRestfulInputRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"required": [],
		"properties": {
			"state_token": { "type": "string" }
		},
		"oneOf": [
			{
				"properties": {
					"input": {
						"type": "object"
					}
				},
				"required": ["input"]
			},
			{
				"properties": {
					"batch_input": {
						"type": "array",
						"items": {
							"type": "object"
						},
						"minItems": 1
					}
				},
				"required": ["batch_input"]
			}
		]
	}
`)

type AuthenticationFlowV1NonRestfulInputRequest struct {
	StateToken string            `json:"state_token,omitempty"`
	Input      json.RawMessage   `json:"input,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

type AuthenticationFlowV1InputHandler struct {
	LoggerFactory *log.Factory
	RedisHandle   *appredis.Handle
	JSON          JSONResponseWriter
	Cookies       AuthenticationFlowV1CookieManager
	Workflows     AuthenticationFlowV1WorkflowService
}

func (h *AuthenticationFlowV1InputHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AuthenticationFlowV1NonRestfulInputRequest
	err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulInputRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	ctx := r.Context()
	if request.Input != nil {
		h.input(ctx, w, r, request)
	} else {
		h.batchInput(ctx, w, r, request)
	}
}

func (h *AuthenticationFlowV1InputHandler) input(ctx context.Context, w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulInputRequest) {
	stateToken := request.StateToken

	output, err := h.input0(ctx, w, r, stateToken, request)
	if err != nil {
		apiResp, apiRespErr := prepareErrorResponse(ctx, h.Workflows, stateToken, err)
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

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1InputHandler) input0(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	stateToken string,
	request AuthenticationFlowV1NonRestfulInputRequest,
) (*authflow.ServiceOutput, error) {
	output, err := h.Workflows.FeedInput(ctx, stateToken, request.Input)
	if err != nil && !errors.Is(err, authflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *AuthenticationFlowV1InputHandler) batchInput(ctx context.Context, w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulInputRequest) {
	stateToken := request.StateToken

	output, err := batchInput0(ctx, h.Workflows, w, r, stateToken, request.BatchInput)
	if err != nil {
		apiResp, apiRespErr := prepareErrorResponse(ctx, h.Workflows, stateToken, err)
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

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}
