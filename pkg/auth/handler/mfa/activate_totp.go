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
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachActivateTOTPHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/totp/activate").
		Handler(server.FactoryToHandler(&ActivateTOTPHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type ActivateTOTPHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ActivateTOTPHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ActivateTOTPHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	TxContext            db.TxContext            `dependency:"TxContext"`
	Validator            *validation.Validator   `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *ActivateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
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
		userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
			MFAOption: authnsession.ResolveMFAOptionOnlyWhenNoAuthenticators,
		})
		if err != nil {
			return err
		}
		recoveryCodes, err := h.MFAProvider.ActivateTOTP(userID, payload.OTP)
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
