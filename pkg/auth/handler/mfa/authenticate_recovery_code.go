package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthenticateRecoveryCodeHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/recovery_code/authenticate").
		Handler(server.FactoryToHandler(&AuthenticateRecoveryCodeHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type AuthenticateRecoveryCodeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthenticateRecoveryCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthenticateRecoveryCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	TxContext            db.TxContext            `dependency:"TxContext"`
	Validator            *validation.Validator   `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz    `dependency:"RequireAuthz"`
	SessionProvider      session.Provider        `dependency:"SessionProvider"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	HookProvider         hook.Provider           `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *AuthenticateRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
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
		h.AuthnSessionProvider.WriteResponse(w, nil, err)
		return
	}

	h.TxContext.UseHook(h.HookProvider)
	result, err := handler.Transactional(h.TxContext, func() (interface{}, error) {
		return h.Handle(payload)
	})
	h.AuthnSessionProvider.WriteResponse(w, result, err)
}

func (h *AuthenticateRecoveryCodeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(AuthenticateRecoveryCodeRequest)

	userID, sess, authnSess, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
		MFAOption: authnsession.ResolveMFAOptionAlwaysAccept,
	})
	if err != nil {
		return
	}

	a, err := h.MFAProvider.AuthenticateRecoveryCode(userID, payload.Code)
	if err != nil {
		return
	}

	opts := coreAuth.AuthnSessionStepMFAOptions{
		AuthenticatorID:   a.ID,
		AuthenticatorType: a.Type,
	}

	if sess != nil {
		err = h.SessionProvider.UpdateMFA(sess, opts)
		if err != nil {
			return
		}
		resp, err = h.AuthnSessionProvider.GenerateResponseWithSession(sess, "")
		if err != nil {
			return
		}
	} else if authnSess != nil {
		err = h.MFAProvider.StepMFA(authnSess, opts)
		if err != nil {
			return
		}
		resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*authnSess)
		if err != nil {
			return
		}
	}

	return
}
