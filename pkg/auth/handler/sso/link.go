package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
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

type OAuthLinkInteractionFlow interface {
	LinkWithOAuthProvider(
		clientID string, userID string, oauthAuthInfo sso.AuthInfo, codeChallenge string,
	) (string, error)

	ExchangeCode(codeHash string, verifier string) (*interactionflows.AuthResult, error)
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
	OAuthProvider sso.OAuthProvider
	Interactions  OAuthLinkInteractionFlow
}

func (h LinkHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h LinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	result.WriteResponse(w)
}

func (h LinkHandler) Handle(w http.ResponseWriter, r *http.Request) (*interactionflows.AuthResult, error) {
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

	var result *interactionflows.AuthResult
	err := db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
		oauthAuthInfo, err := provider.ExternalAccessTokenGetAuthInfo(sso.NewBearerAccessTokenResp(payload.AccessToken))
		if err != nil {
			return err
		}

		code, err := h.Interactions.LinkWithOAuthProvider(
			coreauth.GetAccessKey(r.Context()).Client.ClientID(),
			userID,
			oauthAuthInfo,
			"",
		)
		if err != nil {
			return err
		}

		result, err = h.Interactions.ExchangeCode(code, "")
		if err != nil {
			return err
		}
		return nil
	})
	return result, err
}
