package sso

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/login", &LoginHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderID)
	return handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext)
}

func (f LoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	AccessToken     string              `json:"access_token"`
	MergeRealm      string              `json:"merge_realm"`
	OnUserDuplicate sso.OnUserDuplicate `json:"on_user_duplicate"`
}

// @JSONSchema
const LoginRequestSchema = `
{
	"$id": "#SSOLoginRequest",
	"type": "object",
	"properties": {
		"access_token": { "type": "string" },
		"merge_realm": { "type": "string" },
		"on_user_duplicate": { "type": "string" }
	}
}
`

// Validate request payload
func (p LoginRequestPayload) Validate() (err error) {
	if p.AccessToken == "" {
		err = skyerr.NewInvalidArgument("empty access token", []string{"access_token"})
		return
	}

	if !sso.IsValidOnUserDuplicate(p.OnUserDuplicate) {
		err = skyerr.NewInvalidArgument("Invalid OnUserDuplicate", []string{"on_user_duplicate"})
		return
	}

	return
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
	TxContext           db.TxContext               `dependency:"TxContext"`
	AuthContext         coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	OAuthAuthProvider   oauth.Provider             `dependency:"OAuthAuthProvider"`
	IdentityProvider    principal.IdentityProvider `dependency:"IdentityProvider"`
	AuthInfoStore       authinfo.Store             `dependency:"AuthInfoStore"`
	TokenStore          authtoken.Store            `dependency:"TokenStore"`
	ProviderFactory     *sso.ProviderFactory       `dependency:"SSOProviderFactory"`
	UserProfileStore    userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider        hook.Provider              `dependency:"HookProvider"`
	OAuthConfiguration  config.OAuthConfiguration  `dependency:"OAuthConfiguration"`
	WelcomeEmailEnabled bool                       `dependency:"WelcomeEmailEnabled"`
	TaskQueue           async.Queue                `dependency:"AsyncTaskQueue"`
	Provider            sso.OAuthProvider
	ProviderID          string
}

func (h LoginHandler) WithTx() bool {
	return true
}

func (h LoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LoginRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	if err != nil {
		return payload, err
	}
	if payload.MergeRealm == "" {
		payload.MergeRealm = password.DefaultRealm
	}
	if payload.OnUserDuplicate == "" {
		payload.OnUserDuplicate = sso.OnUserDuplicateDefault
	}
	return payload, nil
}

func (h LoginHandler) Handle(req interface{}) (resp interface{}, err error) {
	if !h.OAuthConfiguration.ExternalAccessTokenFlowEnabled {
		err = skyerr.NewError(skyerr.UndefinedOperation, "External access token flow is disabled")
		return
	}

	provider, ok := h.Provider.(sso.ExternalAccessTokenFlowProvider)
	if !ok {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderID})
		return
	}

	payload := req.(LoginRequestPayload)

	loginState := sso.LoginState{
		MergeRealm:      payload.MergeRealm,
		OnUserDuplicate: payload.OnUserDuplicate,
	}

	oauthAuthInfo, err := provider.ExternalAccessTokenGetAuthInfo(sso.NewBearerAccessTokenResp(payload.AccessToken))
	if err != nil {
		return
	}

	handler := respHandler{
		TokenStore:          h.TokenStore,
		AuthInfoStore:       h.AuthInfoStore,
		OAuthAuthProvider:   h.OAuthAuthProvider,
		IdentityProvider:    h.IdentityProvider,
		UserProfileStore:    h.UserProfileStore,
		HookProvider:        h.HookProvider,
		WelcomeEmailEnabled: h.WelcomeEmailEnabled,
		TaskQueue:           h.TaskQueue,
	}
	resp, err = handler.loginActionResp(oauthAuthInfo, loginState)

	return
}
