package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate mockgen -source=accountmanagement_v1_identification_oauth.go -destination=accountmanagement_v1_identification_oauth_mock_test.go -package api

func ConfigureAccountManagementV1IdentificationOAuthRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/account/identification/oauth")
}

var AccountManagementV1IdentificationOAuthSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"token": {
				"type": "string",
				"minLength": 1
			},
			"query": {
				"type": "string",
				"minLength": 1
			}
		},
		"required": ["token", "query"]
	}
`)

type AccountManagementV1IdentificationOAuthRequest struct {
	Token string `json:"token,omitempty"`
	Query string `json:"query,omitempty"`
}

type AccountManagementV1IdentificationOAuthHandlerService interface {
	FinishAdding(input *accountmanagement.FinishAddingInput) (*accountmanagement.FinishAddingOutput, error)
}

type AccountManagementV1IdentificationOAuthHandler struct {
	JSON    JSONResponseWriter
	Service AccountManagementV1IdentificationOAuthHandlerService
}

func (h *AccountManagementV1IdentificationOAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AccountManagementV1IdentificationOAuthRequest
	err = httputil.BindJSONBody(r, w, AccountManagementV1IdentificationOAuthSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
	h.handle(w, r, request)
}

func (h *AccountManagementV1IdentificationOAuthHandler) handle(w http.ResponseWriter, r *http.Request, request AccountManagementV1IdentificationOAuthRequest) {
	userID := *session.GetUserID(r.Context())
	output, err := h.Service.FinishAdding(&accountmanagement.FinishAddingInput{
		UserID: userID,
		Token:  request.Token,
		Query:  request.Query,
	})
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	h.JSON.WriteResponse(w, &api.Response{Result: output})
}
