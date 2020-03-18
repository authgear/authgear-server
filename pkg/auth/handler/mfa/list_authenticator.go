package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachListAuthenticatorHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/authenticator/list").
		Handler(pkg.MakeHandler(authDependency, newListAuthenticatorHandler)).
		Methods("OPTIONS", "POST")
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
	TxContext     db.TxContext
	Validator     *validation.Validator
	MFAProvider   mfa.Provider
	authnResolver authnResolver
}

func (h *ListAuthenticatorHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
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
		var session coreauthn.Attributer = auth.GetSession(r.Context())
		if session == nil {
			session, err = h.authnResolver.Resolve(
				coreAuth.GetAccessKey(r.Context()).Client,
				payload.AuthnSessionToken,
				authn.SessionStep.IsMFA,
			)
			if err != nil {
				return err
			}
		}

		authenticators, err := h.MFAProvider.ListAuthenticators(session.AuthnAttrs().UserID)
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
