package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachRefreshHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/refresh", &RefreshHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
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
	}
}
`

func (p RefreshRequestPayload) Validate() error {
	if p.RefreshToken == "" {
		return skyerr.NewInvalidArgument("invalid refresh token", []string{"refresh_token"})
	}

	return nil
}

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
	RequireAuthz    handler.RequireAuthz `dependency:"RequireAuthz"`
	TxContext       db.TxContext         `dependency:"TxContext"`
	SessionProvider session.Provider     `dependency:"SessionProvider"`
	SessionWriter   session.Writer       `dependency:"SessionWriter"`
}

func (h RefreshHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
	)
}

func (h RefreshHandler) WithTx() bool {
	return true
}

func (h RefreshHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (RefreshRequestPayload, error) {
	payload := RefreshRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h RefreshHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var result interface{}
	var err error
	defer func() {
		if err == nil {
			refreshResp := result.(RefreshResponse)
			h.SessionWriter.WriteSession(resp, &refreshResp.AccessToken, nil)
			handler.WriteResponse(resp, handler.APIResponse{Result: refreshResp})
		} else {
			handler.WriteResponse(resp, handler.APIResponse{Err: skyerr.MakeError(err)})
		}
	}()

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		return
	}

	result, err = handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		session, err := h.SessionProvider.GetByToken(payload.RefreshToken, coreAuth.SessionTokenKindRefreshToken)
		if err != nil {
			err = skyerr.NewNotAuthenticatedErr()
			return
		}

		err = h.SessionProvider.Refresh(session)
		if err != nil {
			err = skyerr.NewNotAuthenticatedErr()
			return
		}

		result = RefreshResponse{AccessToken: session.AccessToken}
		return
	})
}
