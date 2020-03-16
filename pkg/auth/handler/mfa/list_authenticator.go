package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

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
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachListAuthenticatorHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/authenticator/list").
		Handler(server.FactoryToHandler(&ListAuthenticatorHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type ListAuthenticatorHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ListAuthenticatorHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ListAuthenticatorHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
		"authn_session_token": { "type": "string", "minLength": 1 }
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

		@RequestBody
			@JSONSchema {ListAuthenticatorRequest}
		@Response 200
			List of recovery codes.
			@JSONSchema {ListAuthenticatorResponse}
*/
type ListAuthenticatorHandler struct {
	TxContext            db.TxContext           `dependency:"TxContext"`
	Validator            *validation.Validator  `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz   `dependency:"RequireAuthz"`
	MFAProvider          mfa.Provider           `dependency:"MFAProvider"`
	AuthnSessionProvider authnsession.Provider  `dependency:"AuthnSessionProvider"`
}

func (h *ListAuthenticatorHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *ListAuthenticatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ListAuthenticatorHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ListAuthenticatorRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#ListAuthenticatorRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
			MFAOption: authnsession.ResolveMFAOptionAlwaysAccept,
		})
		if err != nil {
			return err
		}
		authenticators, err := h.MFAProvider.ListAuthenticators(userID)
		if err != nil {
			return err
		}
		resp = ListAuthenticatorResponse{
			Authenticators: authenticators,
		}
		return nil
	})
	return
}
