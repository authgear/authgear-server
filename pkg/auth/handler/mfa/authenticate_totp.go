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
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthenticateTOTPHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/totp/authenticate").
		Handler(auth.MakeHandler(authDependency, newAuthenticateTOTPHandler)).
		Methods("OPTIONS", "POST")
}

type AuthenticateTOTPRequest struct {
	AuthnSessionToken  string `json:"authn_session_token"`
	OTP                string `json:"otp"`
	RequestBearerToken bool   `json:"request_bearer_token"`
}

// @JSONSchema
const AuthenticateTOTPRequestSchema = `
{
	"$id": "#AuthenticateTOTPRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string", "minLength": 1 },
		"otp": { "type": "string", "minLength": 1 },
		"request_bearer_token": { "type": "boolean" }
	},
	"required": ["otp"]
}
`

/*
	@Operation POST /mfa/totp/authenticate - Authenticate with TOTP authenticator.
		Authenticate with TOTP authenticator.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody
			@JSONSchema {AuthenticateTOTPRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateTOTPHandler struct {
	TxContext     db.TxContext
	Validator     *validation.Validator
	TimeProvider  time.Provider
	MFAProvider   mfa.Provider
	authnResolver authnResolver
	authnStepper  authnStepper
}

func (h *AuthenticateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h *AuthenticateTOTPHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (AuthenticateTOTPRequest, error) {
	payload := AuthenticateTOTPRequest{}
	err := handler.BindJSONBody(request, resp, h.Validator, "#AuthenticateTOTPRequest", &payload)
	return payload, err
}

func (h *AuthenticateTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	payload, err := h.DecodeRequest(r, w)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	var result authn.Result
	err = db.WithTx(h.TxContext, func() error {
		var session coreauthn.Attributer = coreauthn.GetSession(r.Context())
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

		attrs := session.AuthnAttrs()
		a, bearerToken, err := h.MFAProvider.AuthenticateTOTP(
			attrs.UserID,
			payload.OTP,
			payload.RequestBearerToken,
		)
		if err != nil {
			return err
		}

		now := h.TimeProvider.NowUTC()
		attrs.AuthenticatorID = a.ID
		attrs.AuthenticatorType = a.Type
		attrs.AuthenticatorUpdatedAt = &now

		result, err = h.authnStepper.StepSession(
			coreAuth.GetAccessKey(r.Context()).Client,
			session,
			bearerToken,
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	h.authnStepper.WriteResult(w, result)
}
