package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachLinkHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/link").
		Handler(server.FactoryToHandler(&LinkHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type LinkHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LinkHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LinkHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.OAuthProvider = h.ProviderFactory.NewOAuthProvider(h.ProviderID)
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
	TxContext          db.TxContext              `dependency:"TxContext"`
	Validator          *validation.Validator     `dependency:"Validator"`
	AuthContext        coreAuth.ContextGetter    `dependency:"AuthContextGetter"`
	RequireAuthz       handler.RequireAuthz      `dependency:"RequireAuthz"`
	HookProvider       hook.Provider             `dependency:"HookProvider"`
	ProviderFactory    *sso.OAuthProviderFactory `dependency:"SSOOAuthProviderFactory"`
	SSOProvider        sso.Provider              `dependency:"SSOProvider"`
	AuthnOAuthProvider authn.OAuthProvider       `dependency:"AuthnOAuthProvider"`
	OAuthProvider      sso.OAuthProvider
	ProviderID         string
}

func (h LinkHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
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

	if !h.SSOProvider.IsExternalAccessTokenFlowEnabled() {
		err = skyerr.NewNotFound("external access token flow is disabled")
		return
	}

	provider, ok := h.OAuthProvider.(sso.ExternalAccessTokenFlowProvider)
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

		code, err := h.AuthnOAuthProvider.LinkOAuth(oauthAuthInfo, "", linkState)
		if err != nil {
			return err
		}

		authInfo, userProfile, _, err := h.AuthnOAuthProvider.ExtractAuthorizationCode(code)
		if err != nil {
			return err
		}

		user := model.NewUser(*authInfo, *userProfile)
		resp = model.NewAuthResponseWithUser(user)
		return nil
	})
	return
}
