package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachListAuthenticatorHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/authenticator/list", &ListAuthenticatorHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type ListAuthenticatorHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ListAuthenticatorHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ListAuthenticatorHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

type ListAuthenticatorRequest struct {
	AuthnSessionToken string `json:"authn_session_token"`
}

func (p ListAuthenticatorRequest) Validate() error {
	return nil
}

// @JSONSchema
const ListAuthenticatorRequestSchema = `
{
	"$id": "#ListAuthenticatorRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string" }
	}
}
`

type ListAuthenticatorResponse struct {
	Authenticators []mfa.Authenticator `json:"authenticators"`
}

// @JSONSchema
const ListAuthenticatorResponseSchema = `
{
	"$id": "#ListAuthenticatorResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"authenticators": {
					"type": "array",
					"items": { "type": "object" }
				}
			}
		}
	}
}
`

/*
	@Operation POST /mfa/authenticator/list - List authenticators
		List authenticators.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody {ListAuthenticatorRequest}
		@Response 200
			List of recovery codes.
			@JSONSchema {ListAuthenticatorResponse}
*/
type ListAuthenticatorHandler struct {
	TxContext            db.TxContext           `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	MFAProvider          mfa.Provider           `dependency:"MFAProvider"`
	AuthnSessionProvider authnsession.Provider  `dependency:"AuthnSessionProvider"`
}

func (h *ListAuthenticatorHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *ListAuthenticatorHandler) WithTx() bool {
	return true
}

func (h *ListAuthenticatorHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := ListAuthenticatorRequest{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h *ListAuthenticatorHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ListAuthenticatorRequest)
	userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
		MFAOption: authnsession.ResolveMFAOptionAlwaysAccept,
	})
	if err != nil {
		return nil, err
	}
	authenticators, err := h.MFAProvider.ListAuthenticators(userID)
	if err != nil {
		return nil, err
	}
	return ListAuthenticatorResponse{
		Authenticators: authenticators,
	}, nil
}
