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

func AttachActivateTOTPHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/totp/activate").
		Handler(pkg.MakeHandler(authDependency, newActivateTOTPHandler)).
		Methods("OPTIONS", "POST")
}

type ActivateTOTPRequest struct {
	OTP               string `json:"otp"`
	AuthnSessionToken string `json:"authn_session_token"`
}

type ActivateTOTPResponse struct {
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

// @JSONSchema
const ActivateTOTPRequestSchema = `
{
	"$id": "#ActivateTOTPRequest",
	"type": "object",
	"properties": {
		"otp": { "type": "string", "minLength": 1 },
		"authn_session_token": { "type": "string", "minLength": 1 }
	},
	"required": ["otp"]
}
`

// @JSONSchema
const ActivateTOTPResponseSchema = `
{
	"$id": "#ActivateTOTPResponse",
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
	@Operation POST /mfa/totp/activate - Activate TOTP authenticator.
		Activate TOTP authenticator.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {ActivateTOTPRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {ActivateTOTPResponse}
*/
type ActivateTOTPHandler struct {
	TxContext     db.TxContext
	Validator     *validation.Validator
	MFAProvider   mfa.Provider
	authnResolver authnResolver
}

func (h *ActivateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h *ActivateTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ActivateTOTPHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ActivateTOTPRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#ActivateTOTPRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		var session coreauthn.Attributer = auth.GetSession(r.Context())
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

		recoveryCodes, err := h.MFAProvider.ActivateTOTP(session.AuthnAttrs().UserID, payload.OTP)
		if err != nil {
			return err
		}

		resp = ActivateTOTPResponse{
			RecoveryCodes: recoveryCodes,
		}
		return nil
	})
	return
}
