package sso

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
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

func AttachLinkHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/link", &LinkHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type LinkHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LinkHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LinkHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderID)
	return h.RequireAuthz(h, h)
}

// LinkRequestPayload login handler request payload
type LinkRequestPayload struct {
	AccessToken string `json:"access_token"`
}

// @JSONSchema
const LinkRequestSchema = `
{
	"$id": "#SSOLinkRequest",
	"type": "object",
	"properties": {
		"access_token": { "type": "string", "minLength": 1 }
	},
	"required": ["access_token"]
}
`

/*
	@Operation POST /sso/{provider_id}/link - Link SSO provider with token
		Link the specified SSO provider with the current user, using access
		token obtained from the provider.

		@Tag SSO
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Parameter {SSOProviderID}
		@RequestBody
			Describe the access token of SSO provider.
			@JSONSchema {SSOLinkRequest}
		@Response 200 {EmptyResponse}

		@Callback identity_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type LinkHandler struct {
	TxContext          db.TxContext               `dependency:"TxContext"`
	Validator          *validation.Validator      `dependency:"Validator"`
	AuthContext        coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz       handler.RequireAuthz       `dependency:"RequireAuthz"`
	OAuthAuthProvider  oauth.Provider             `dependency:"OAuthAuthProvider"`
	IdentityProvider   principal.IdentityProvider `dependency:"IdentityProvider"`
	AuthInfoStore      authinfo.Store             `dependency:"AuthInfoStore"`
	UserProfileStore   userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider       hook.Provider              `dependency:"HookProvider"`
	ProviderFactory    *sso.ProviderFactory       `dependency:"SSOProviderFactory"`
	OAuthConfiguration config.OAuthConfiguration  `dependency:"OAuthConfiguration"`
	Provider           sso.OAuthProvider
	ProviderID         string
}

func (h LinkHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
}

func (h LinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h LinkHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload LinkRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#SSOLinkRequest", &payload); err != nil {
		return nil, err
	}

	if !h.OAuthConfiguration.ExternalAccessTokenFlowEnabled {
		err = skyerr.NewNotFound("external access token flow is disabled")
		return
	}

	provider, ok := h.Provider.(sso.ExternalAccessTokenFlowProvider)
	if !ok {
		err = skyerr.NewNotFound("unknown provider")
		return
	}

	err = hook.WithTx(h.HookProvider, h.TxContext, func() error {
		authInfo, _ := h.AuthContext.AuthInfo()
		userID := authInfo.ID

		linkState := sso.LinkState{
			UserID: userID,
		}
		oauthAuthInfo, err := provider.ExternalAccessTokenGetAuthInfo(sso.NewBearerAccessTokenResp(payload.AccessToken))
		if err != nil {
			return err
		}

		handler := respHandler{
			AuthInfoStore:     h.AuthInfoStore,
			OAuthAuthProvider: h.OAuthAuthProvider,
			IdentityProvider:  h.IdentityProvider,
			UserProfileStore:  h.UserProfileStore,
			HookProvider:      h.HookProvider,
		}
		resp, err = handler.linkActionResp(oauthAuthInfo, linkState)
		return err
	})
	return
}
