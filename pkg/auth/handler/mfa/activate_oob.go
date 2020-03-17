package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachActivateOOBHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/oob/activate").
		Handler(auth.MakeHandler(authDependency, newActivateOOBHandler)).
		Methods("OPTIONS", "POST")
}

type ActivateOOBRequest struct {
	Code              string `json:"code"`
	AuthnSessionToken string `json:"authn_session_token"`
}
type ActivateOOBResponse struct {
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

// @JSONSchema
const ActivateOOBRequestSchema = `
{
	"$id": "#ActivateOOBRequest",
	"type": "object",
	"properties": {
		"code": { "type": "string", "minLength": 1 },
		"authn_session_token": { "type": "string", "minLength": 1 }
	},
	"required": ["code"]
}
`

// @JSONSchema
const ActivateOOBResponseSchema = `
{
	"$id": "#ActivateOOBResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"recovery_codes": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			}
		}
	}
}
`

/*
	@Operation POST /mfa/oob/activate - Activate OOB authenticator.
		Activate OOB authenticator.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {ActivateOOBRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {ActivateOOBResponse}
*/
type ActivateOOBHandler struct {
	TxContext     db.TxContext
	Validator     *validation.Validator
	MFAProvider   mfa.Provider
	authnResolver authnResolver
}

func (h *ActivateOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *ActivateOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ActivateOOBHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ActivateOOBRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#ActivateOOBRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		session := authn.GetSession(r.Context())
		if session == nil {
			session, err = h.authnResolver.Resolve(
				coreAuth.GetAccessKey(r.Context()).Client,
				payload.AuthnSessionToken,
				func(s authn.SessionStep) bool { return s == authn.SessionStepMFASetup },
			)
			if err != nil {
				return err
			}
		}

		recoveryCodes, err := h.MFAProvider.ActivateOOB(session.SessionAttrs().UserID, payload.Code)
		if err != nil {
			return err
		}

		resp = ActivateOOBResponse{
			RecoveryCodes: recoveryCodes,
		}
		return nil
	})
	return
}
