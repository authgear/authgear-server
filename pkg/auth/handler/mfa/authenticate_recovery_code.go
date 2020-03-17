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

func AttachAuthenticateRecoveryCodeHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/recovery_code/authenticate").
		Handler(auth.MakeHandler(authDependency, newAuthenticateRecoveryCodeHandler)).
		Methods("OPTIONS", "POST")
}

type AuthenticateRecoveryCodeRequest struct {
	AuthnSessionToken string `json:"authn_session_token"`
	Code              string `json:"code"`
}

// @JSONSchema
const AuthenticateRecoveryCodeRequestSchema = `
{
	"$id": "#AuthenticateRecoveryCodeRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string", "minLength": 1 },
		"code": { "type": "string", "minLength": 1 }
	},
	"required": ["code"]
}
`

/*
	@Operation POST /mfa/recovery_code/authenticate - Authenticate with recovery code.
		Authenticate with recovery code.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody
			@JSONSchema {AuthenticateRecoveryCodeRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateRecoveryCodeHandler struct {
	TxContext     db.TxContext
	Validator     *validation.Validator
	TimeProvider  time.Provider
	MFAProvider   mfa.Provider
	authnResolver authnResolver
	authnStepper  authnStepper
}

func (h *AuthenticateRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h *AuthenticateRecoveryCodeHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (AuthenticateRecoveryCodeRequest, error) {
	payload := AuthenticateRecoveryCodeRequest{}
	err := handler.BindJSONBody(request, resp, h.Validator, "#AuthenticateRecoveryCodeRequest", &payload)
	return payload, err
}

func (h *AuthenticateRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		a, err := h.MFAProvider.AuthenticateRecoveryCode(
			attrs.UserID,
			payload.Code,
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
			"",
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
