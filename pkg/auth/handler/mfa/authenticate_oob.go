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

func AttachAuthenticateOOBHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/oob/authenticate").
		Handler(server.FactoryToHandler(&AuthenticateOOBHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type AuthenticateOOBHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthenticateOOBHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthenticateOOBHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type AuthenticateOOBRequest struct {
	AuthnSessionToken  string `json:"authn_session_token"`
	Code               string `json:"code"`
	RequestBearerToken bool   `json:"request_bearer_token"`
}

// @JSONSchema
const AuthenticateOOBRequestSchema = `
{
	"$id": "#AuthenticateOOBRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string", "minLength": 1 },
		"code": { "type": "string", "minLength": 1 },
		"request_bearer_token": { "type": "boolean" }
	},
	"required": ["code"]
}
`

/*
	@Operation POST /mfa/oob/authenticate - Authenticate with OOB authenticator.
		Authenticate with OOB authenticator.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody
			@JSONSchema {AuthenticateOOBRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateOOBHandler struct {
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

func (h *AuthenticateOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireClient),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *AuthenticateOOBHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (AuthenticateOOBRequest, error) {
	payload := AuthenticateOOBRequest{}
	err := handler.BindJSONBody(request, resp, h.Validator, "#AuthenticateOOBRequest", &payload)
	return payload, err
}

func (h *AuthenticateOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	payload, err := h.DecodeRequest(r, w)
	if err != nil {
		h.AuthnSessionProvider.WriteResponse(w, nil, err)
		return
	}

	result, err := handler.Transactional(h.TxContext, func() (interface{}, error) {
		return h.Handle(payload)
	})
	h.AuthnSessionProvider.WriteResponse(w, result, err)
}

func (h *AuthenticateOOBHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(AuthenticateOOBRequest)

	userID, sess, authnSess, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
		MFAOption: authnsession.ResolveMFAOptionAlwaysAccept,
	})
	if err != nil {
		return
	}

	a, bearerToken, err := h.MFAProvider.AuthenticateOOB(userID, payload.Code, payload.RequestBearerToken)
	if err != nil {
		return
	}
	opts := coreAuth.AuthnSessionStepMFAOptions{
		AuthenticatorID:          a.ID,
		AuthenticatorType:        a.Type,
		AuthenticatorOOBChannel:  a.Channel,
		AuthenticatorBearerToken: bearerToken,
	}

	if sess != nil {
		err = h.SessionProvider.UpdateMFA(sess, opts)
		if err != nil {
			return
		}
		resp, err = h.AuthnSessionProvider.GenerateResponseWithSession(sess, bearerToken)
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
