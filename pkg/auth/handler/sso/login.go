package sso

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachLoginHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/login").
		Handler(server.FactoryToHandler(&LoginHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.OAuthProvider = h.ProviderFactory.NewOAuthProvider(h.ProviderID)
	return h.RequireAuthz(h, h)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	AccessToken     string                `json:"access_token"`
	MergeRealm      string                `json:"-"`
	OnUserDuplicate model.OnUserDuplicate `json:"on_user_duplicate"`
}

func (p *LoginRequestPayload) SetDefaultValue() {
	if p.MergeRealm == "" {
		p.MergeRealm = password.DefaultRealm
	}
	if p.OnUserDuplicate == "" {
		p.OnUserDuplicate = model.OnUserDuplicateDefault
	}
}

// @JSONSchema
const LoginRequestSchema = `
{
	"$id": "#SSOLoginRequest",
	"type": "object",
	"properties": {
		"access_token": { "type": "string", "minLength": 1 },
		"on_user_duplicate": {"type": "string", "enum": ["abort", "merge", "create"] }
	},
	"required": ["access_token"]
}
`

type LoginAuthnProvider interface {
	OAuthAuthenticate(
		authInfo sso.AuthInfo,
		codeChallenge string,
		loginState sso.LoginState,
	) (*sso.SkygearAuthorizationCode, error)

	OAuthExchangeCode(
		client config.OAuthClientConfiguration,
		session *session.Session,
		code *sso.SkygearAuthorizationCode,
	) (authn.Result, error)
}

/*
	@Operation POST /sso/{provider_id}/login - Login SSO provider with token
		Login the specified SSO provider, using access token obtained from the provider.

		@Tag SSO

		@Parameter {SSOProviderID}
		@RequestBody
			Describe the access token of SSO provider and login behavior.
			@JSONSchema {SSOLoginRequest}
		@Response 200 {EmptyResponse}

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type LoginHandler struct {
	TxContext       db.TxContext              `dependency:"TxContext"`
	Validator       *validation.Validator     `dependency:"Validator"`
	RequireAuthz    handler.RequireAuthz      `dependency:"RequireAuthz"`
	ProviderFactory *sso.OAuthProviderFactory `dependency:"SSOOAuthProviderFactory"`
	AuthnProvider   LoginAuthnProvider        `dependency:"AuthnProvider"`
	SSOProvider     sso.Provider              `dependency:"SSOProvider"`
	OAuthProvider   sso.OAuthProvider
	ProviderID      string
}

func (h LoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h LoginHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload LoginRequestPayload, err error) {
	err = handler.BindJSONBody(request, resp, h.Validator, "#SSOLoginRequest", &payload)
	return
}

func (h LoginHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	var result authn.Result
	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Handle(req, payload)
		return
	})
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	// TODO(authn): write response
	fmt.Printf("%#v\n", result)
}

func (h LoginHandler) Handle(r *http.Request, payload LoginRequestPayload) (authn.Result, error) {
	if !h.SSOProvider.IsExternalAccessTokenFlowEnabled() {
		return nil, skyerr.NewNotFound("external access token flow is disabled")
	}

	provider, ok := h.OAuthProvider.(sso.ExternalAccessTokenFlowProvider)
	if !ok {
		return nil, skyerr.NewNotFound("unknown provider")
	}

	loginState := sso.LoginState{
		MergeRealm:      payload.MergeRealm,
		OnUserDuplicate: payload.OnUserDuplicate,
	}

	oauthAuthInfo, err := provider.ExternalAccessTokenGetAuthInfo(sso.NewBearerAccessTokenResp(payload.AccessToken))
	if err != nil {
		return nil, err
	}

	code, err := h.AuthnProvider.OAuthAuthenticate(oauthAuthInfo, "", loginState)
	if err != nil {
		return nil, err
	}

	result, err := h.AuthnProvider.OAuthExchangeCode(
		coreauth.GetAccessKey(r.Context()).Client,
		nil, // Assume no session for SSO login
		code,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
