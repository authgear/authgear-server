package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/apiclientconfig"
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
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachAuthenticateBearerTokenHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/bearer_token/authenticate", &AuthenticateBearerTokenHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type AuthenticateBearerTokenHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthenticateBearerTokenHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthenticateBearerTokenHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(h, h.AuthContext, h)
}

type AuthenticateBearerTokenRequest struct {
	AuthnSessionToken string `json:"authn_session_token"`
	BearerToken       string `json:"bearer_token"`
}

func (r AuthenticateBearerTokenRequest) Validate() error {
	if r.BearerToken == "" {
		return skyerr.NewInvalidArgument("missing bearer token", []string{"bearer_token"})
	}
	return nil
}

// nolint: gosec
// @JSONSchema
const AuthenticateBearerTokenRequestSchema = `
{
	"$id": "#AuthenticateBearerTokenRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string" },
		"code": { "type": "string" }
	},
	"required": ["code"]
}
`

/*
	@Operation POST /mfa/bearer_token/authenticate - Authenticate with bearer token.
		Authenticate with bearer token.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody {AuthenticateBearerTokenRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateBearerTokenHandler struct {
	TxContext                      db.TxContext             `dependency:"TxContext"`
	AuthContext                    coreAuth.ContextGetter   `dependency:"AuthContextGetter"`
	SessionProvider                session.Provider         `dependency:"SessionProvider"`
	MFAProvider                    mfa.Provider             `dependency:"MFAProvider"`
	MFAConfiguration               config.MFAConfiguration  `dependency:"MFAConfiguration"`
	HookProvider                   hook.Provider            `dependency:"HookProvider"`
	AuthnSessionProvider           authnsession.Provider    `dependency:"AuthnSessionProvider"`
	APIClientConfigurationProvider apiclientconfig.Provider `dependency:"APIClientConfigurationProvider"`
}

func (h *AuthenticateBearerTokenHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

func (h *AuthenticateBearerTokenHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthenticateBearerTokenRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	if err != nil {
		return nil, err
	}

	apiClientConfig, ok := h.APIClientConfigurationProvider.Get()
	if ok && apiClientConfig.SessionTransport == config.SessionTransportTypeCookie {
		cookie, err := request.Cookie(coreHttp.CookieNameMFABearerToken)
		if err == nil {
			payload.BearerToken = cookie.Value
		}
	}

	return payload, nil
}

func (h *AuthenticateBearerTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *AuthenticateBearerTokenHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(AuthenticateBearerTokenRequest)

	userID, sess, authnSess, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
		MFAOption: authnsession.ResolveMFAOptionAlwaysAccept,
	})
	if err != nil {
		return
	}

	a, err := h.MFAProvider.AuthenticateBearerToken(userID, payload.BearerToken)
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
