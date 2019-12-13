package sso

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAppleResponseHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/apple_response", &AppleResponseHandlerFactory{
		Dependency: authDependency,
	})
	return server
}

type AppleResponseHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f *AppleResponseHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AppleResponseHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.OAuthProvider = h.ProviderFactory.NewOAuthProvider(h.ProviderID)
	return h.RequireAuthz(h, h)
}

type AppleResponseRequest struct {
	AuthorizationCode string `json:"authorization_code"`
	State             string `json:"state"`
	Nonce             string `json:"nonce"`
	Scope             string `json:"scope"`
	CodeVerifier      string `json:"code_verifier"`
}

// @JSONSchema
const AppleResponseRequestSchema = `
{
	"$id": "#AppleResponseRequest",
	"type": "object",
	"properties": {
		"authorization_code": { "type": "string", "minLength": 1 },
		"state": { "type": "string", "minLength": 1 },
		"nonce": { "type": "string", "minLength": 1 },
		"scope": { "type": "string" },
		"code_verifier": { "type": "string", "minLength": 1 }
	},
	"required": ["authorization_code", "state", "nonce", "scope", "code_verifier"]
}
`

type AppleResponseHandler struct {
	RequireAuthz         handler.RequireAuthz       `dependency:"RequireAuthz"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	HookProvider         hook.Provider              `dependency:"HookProvider"`
	Validator            *validation.Validator      `dependency:"Validator"`
	AuthnSessionProvider authnsession.Provider      `dependency:"AuthnSessionProvider"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	OAuthAuthProvider    oauth.Provider             `dependency:"OAuthAuthProvider"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
	ProviderFactory      *sso.OAuthProviderFactory  `dependency:"SSOOAuthProviderFactory"`
	SSOProvider          sso.Provider               `dependency:"SSOProvider"`
	TaskQueue            async.Queue                `dependency:"AsyncTaskQueue"`
	WelcomeEmailEnabled  bool                       `dependency:"WelcomeEmailEnabled"`
	URLPrefix            *url.URL                   `dependency:"URLPrefix"`
	OAuthProvider        sso.OAuthProvider
	ProviderID           string
}

func (h *AppleResponseHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *AppleResponseHandler) DecodeRequest(w http.ResponseWriter, r *http.Request) (payload *AppleResponseRequest, err error) {
	err = handler.BindJSONBody(r, w, h.Validator, "#AppleResponseRequest", &payload)
	return
}

func (h *AppleResponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var result interface{}
	var err error

	payload, err := h.DecodeRequest(w, r)
	if err != nil {
		h.AuthnSessionProvider.WriteResponse(w, nil, err)
		return
	}

	err = hook.WithTx(h.HookProvider, h.TxContext, func() (err error) {
		result, err = h.Handle(payload)
		return
	})
	h.AuthnSessionProvider.WriteResponse(w, result, err)
}

func (h *AppleResponseHandler) Handle(payload *AppleResponseRequest) (result interface{}, err error) {
	if h.OAuthProvider == nil {
		err = skyerr.NewNotFound("unknown provider")
		return
	}

	state, err := h.SSOProvider.DecodeState(payload.State)
	if err != nil {
		return
	}

	oauthAuthInfo, err := h.OAuthProvider.GetAuthInfo(
		sso.OAuthAuthorizationResponse{
			Code:  payload.AuthorizationCode,
			State: payload.State,
			Scope: payload.Scope,
			Nonce: payload.Nonce,
		},
		*state,
	)
	if err != nil {
		return
	}

	respHandler := respHandler{
		AuthnSessionProvider: h.AuthnSessionProvider,
		AuthInfoStore:        h.AuthInfoStore,
		OAuthAuthProvider:    h.OAuthAuthProvider,
		IdentityProvider:     h.IdentityProvider,
		UserProfileStore:     h.UserProfileStore,
		HookProvider:         h.HookProvider,
		WelcomeEmailEnabled:  h.WelcomeEmailEnabled,
		TaskQueue:            h.TaskQueue,
		URLPrefix:            h.URLPrefix,
	}

	var code *sso.SkygearAuthorizationCode
	if state.Action == "login" {
		code, err = respHandler.LoginCode(oauthAuthInfo, state.CodeChallenge, state.LoginState)
	} else {
		code, err = respHandler.LinkCode(oauthAuthInfo, state.CodeChallenge, state.LinkState)
	}
	if err != nil {
		return
	}

	err = h.SSOProvider.VerifyPKCE(code, payload.CodeVerifier)
	if err != nil {
		return
	}

	return respHandler.CodeToResponse(code)
}
