package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachRefreshHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/refresh").
		Handler(auth.MakeHandler(authDependency, newRefreshHandler)).
		Methods("OPTIONS", "POST")
}

type RefreshRequestPayload struct {
	RefreshToken string `json:"refresh_token"`
}

// @JSONSchema
const RefreshRequestSchema = `
{
	"$id": "#RefreshRequest",
	"type": "object",
	"properties": {
		"refresh_token": { "type": "string" }
	},
	"required": ["refresh_token"]
}
`

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

// @JSONSchema
const RefreshResponseSchema = `
{
	"$id": "#RefreshResponse",
	"type": "object",
	"properties": {
		"access_token": { "type": "string" }
	}
}
`

type refreshProvider interface {
	RefreshAPIToken(
		client config.OAuthClientConfiguration,
		refreshToken string,
	) (accessToken string, err error)
}

/*
	@Operation POST /refresh - Refresh access token
		Returns new access token, using the refresh token.

		@Tag User

		@RequestBody
			Describe refresh token of session.
			@JSONSchema {RefreshRequest}

		@Response 200
			New access token.
			@JSONSchema {RefreshResponse}
*/
type RefreshHandler struct {
	validator       *validation.Validator
	txContext       db.TxContext
	refreshProvider refreshProvider
}

func (h RefreshHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
	)
}

func (h RefreshHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := h.Handle(resp, req)
	if err == nil {
		handler.WriteResponse(resp, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
	}
}

func (h RefreshHandler) Handle(resp http.ResponseWriter, req *http.Request) (result RefreshResponse, err error) {
	var payload RefreshRequestPayload
	if err = handler.BindJSONBody(req, resp, h.validator, "#RefreshRequest", &payload); err != nil {
		return
	}

	err = db.WithTx(h.txContext, func() error {
		client := coreAuth.GetAccessKey(req.Context()).Client
		accessToken, err := h.refreshProvider.RefreshAPIToken(client, payload.RefreshToken)
		if err != nil {
			return authz.ErrNotAuthenticated
		}

		result = RefreshResponse{AccessToken: accessToken}
		return nil
	})
	return result, err
}
