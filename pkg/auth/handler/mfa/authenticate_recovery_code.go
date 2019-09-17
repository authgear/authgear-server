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

func AttachAuthenticateRecoveryCodeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/recovery_code/authenticate", &AuthenticateRecoveryCodeHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type AuthenticateRecoveryCodeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthenticateRecoveryCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthenticateRecoveryCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(h, h.AuthContext, h)
}

type AuthenticateRecoveryCodeRequest struct {
	AuthnSessionToken string `json:"authn_session_token"`
	Code              string `json:"code"`
}

func (r AuthenticateRecoveryCodeRequest) Validate() error {
	if r.AuthnSessionToken == "" {
		return skyerr.NewInvalidArgument("missing authentication session token", []string{"authn_session_token"})
	}
	if r.Code == "" {
		return skyerr.NewInvalidArgument("missing recovery code", []string{"code"})
	}
	return nil
}

// @JSONSchema
const AuthenticateRecoveryCodeRequestSchema = `
{
	"$id": "#AuthenticateRecoveryCodeRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string" },
		"code": { "type": "string" }
	}
	"required": ["authn_session_token", "code"]
}
`

/*
	@Operation POST /mfa/recovery_code/authenticate - Authenticate with recovery code.
		Authenticate with recovery code.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody {AuthenticateRecoveryCodeRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateRecoveryCodeHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	HookProvider         hook.Provider           `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *AuthenticateRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(authz.PolicyFunc(policy.DenyNoAccessKey))
}

func (h *AuthenticateRecoveryCodeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthenticateRecoveryCodeRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *AuthenticateRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *AuthenticateRecoveryCodeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(AuthenticateRecoveryCodeRequest)

	authnSess, err := h.AuthnSessionProvider.NewWithToken(payload.AuthnSessionToken)
	if err != nil {
		return
	}

	a, err := h.MFAProvider.AuthenticateRecoveryCode(authnSess.UserID, payload.Code)
	if err != nil {
		return
	}

	err = h.MFAProvider.StepMFA(authnSess, coreAuth.AuthnSessionStepMFAOptions{
		AuthenticatorID:   a.ID,
		AuthenticatorType: a.Type,
	})
	if err != nil {
		return
	}

	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*authnSess)
	if err != nil {
		return
	}

	return
}
