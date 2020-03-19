package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachLinkHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/link").
		Handler(pkg.MakeHandler(authDependency, newLinkHandler)).
		Methods("OPTIONS", "POST")
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

type LinkAuthnProvider interface {
	OAuthLink(
		authInfo sso.AuthInfo,
		codeChallenge string,
		linkState sso.LinkState,
	) (*sso.SkygearAuthorizationCode, error)

	OAuthExchangeCode(
		client config.OAuthClientConfiguration,
		session auth.AuthSession,
		code *sso.SkygearAuthorizationCode,
	) (authn.Result, error)

	WriteResult(rw http.ResponseWriter, result authn.Result)
}

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
	TxContext     db.TxContext
	Validator     *validation.Validator
	SSOProvider   sso.Provider
	AuthnProvider LinkAuthnProvider
	OAuthProvider sso.OAuthProvider
}

func (h LinkHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h LinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	h.AuthnProvider.WriteResult(w, result)
}

func (h LinkHandler) Handle(w http.ResponseWriter, r *http.Request) (authn.Result, error) {
	var payload LinkRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#SSOLinkRequest", &payload); err != nil {
		return nil, err
	}

	if !h.SSOProvider.IsExternalAccessTokenFlowEnabled() {
		return nil, skyerr.NewNotFound("external access token flow is disabled")
	}

	provider, ok := h.OAuthProvider.(sso.ExternalAccessTokenFlowProvider)
	if !ok {
		return nil, skyerr.NewNotFound("unknown provider")
	}

	var result authn.Result
	err := db.WithTx(h.TxContext, func() error {
		userID := auth.GetUser(r.Context()).ID

		linkState := sso.LinkState{
			UserID: userID,
		}
		oauthAuthInfo, err := provider.ExternalAccessTokenGetAuthInfo(sso.NewBearerAccessTokenResp(payload.AccessToken))
		if err != nil {
			return err
		}

		code, err := h.AuthnProvider.OAuthLink(oauthAuthInfo, "", linkState)
		if err != nil {
			return err
		}

		result, err = h.AuthnProvider.OAuthExchangeCode(
			coreauth.GetAccessKey(r.Context()).Client,
			auth.GetSession(r.Context()),
			code,
		)
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}
