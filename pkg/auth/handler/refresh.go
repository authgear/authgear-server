package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachRefreshHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/refresh").
		Handler(server.FactoryToHandler(&RefreshHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type RefreshHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RefreshHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RefreshHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	RequireAuthz    handler.RequireAuthz  `dependency:"RequireAuthz"`
	Validator       *validation.Validator `dependency:"Validator"`
	TxContext       db.TxContext          `dependency:"TxContext"`
	SessionProvider session.Provider      `dependency:"SessionProvider"`
	SessionWriter   session.Writer        `dependency:"SessionWriter"`
}

func (h RefreshHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
	)
}

func (h RefreshHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := h.Handle(resp, req)
	if err == nil {
		h.SessionWriter.WriteSession(resp, &result.AccessToken, nil)
		handler.WriteResponse(resp, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
	}
}

func (h RefreshHandler) Handle(resp http.ResponseWriter, req *http.Request) (result RefreshResponse, err error) {
	var payload RefreshRequestPayload
	if err = handler.BindJSONBody(req, resp, h.Validator, "#RefreshRequest", &payload); err != nil {
		return
	}

	err = db.WithTx(h.TxContext, func() error {
		s, err := h.SessionProvider.GetByToken(payload.RefreshToken, coreAuth.SessionTokenKindRefreshToken)
		if err != nil {
			if errors.Is(err, session.ErrSessionNotFound) {
				err = authz.ErrNotAuthenticated
			}
			return err
		}

		accessToken, err := h.SessionProvider.Refresh(s)
		if err != nil {
			return err
		}

		result = RefreshResponse{AccessToken: accessToken}
		return nil
	})
	return result, err
}
