package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachAuthenticateTOTPHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/totp/authenticate", &AuthenticateTOTPHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type AuthenticateTOTPHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthenticateTOTPHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthenticateTOTPHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(h, h.AuthContext, h)
}

type AuthenticateTOTPRequest struct {
	AuthnSessionToken  string `json:"authn_session_token"`
	OTP                string `json:"otp"`
	RequestBearerToken bool   `json:"request_bearer_token"`
}

func (r AuthenticateTOTPRequest) Validate() error {
	if r.AuthnSessionToken == "" {
		return skyerr.NewInvalidArgument("missing authentication session token", []string{"authn_session_token"})
	}
	if r.OTP == "" {
		return skyerr.NewInvalidArgument("missing OTP", []string{"otp"})
	}
	return nil
}

// @JSONSchema
const AuthenticateTOTPRequestSchema = `
{
	"$id": "#AuthenticateTOTPRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string" },
		"otp": { "type": "string" },
		"request_bearer_token": { "type": "boolean" }
	}
	"required": ["authn_session_token", "otp"]
}
`

/*
	@Operation POST /mfa/totp/authenticate - Authenticate with TOTP.
		Authenticate with TOTP.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody {AuthenticateTOTPRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateTOTPHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	HookProvider         hook.Provider           `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *AuthenticateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(authz.PolicyFunc(policy.DenyNoAccessKey))
}

func (h *AuthenticateTOTPHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthenticateTOTPRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *AuthenticateTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	defer func() {
		if err == nil {
			h.HookProvider.DidCommitTx()
		}
		h.AuthnSessionProvider.WriteResponse(w, result, err)
	}()

	payload, err := h.DecodeRequest(r)
	if err != nil {
		return
	}

	err = payload.Validate()
	if err != nil {
		return
	}

	result, err = handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		result, err = h.Handle(payload)
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
}

func (h *AuthenticateTOTPHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(AuthenticateTOTPRequest)

	authnSess, err := h.AuthnSessionProvider.NewWithToken(payload.AuthnSessionToken)
	if err != nil {
		return
	}

	// TODO(mfa): bearer token
	a, err := h.MFAProvider.AuthenticateTOTP(authnSess.UserID, payload.OTP)
	if err != nil {
		return
	}

	err = authnSess.StepMFA(a.ID, a.Type, "")
	if err != nil {
		return
	}

	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*authnSess)
	if err != nil {
		return
	}

	return
}
