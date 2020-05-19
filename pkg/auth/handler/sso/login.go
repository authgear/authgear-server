package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachLoginHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/login").
		Handler(pkg.MakeHandler(authDependency, newLoginHandler)).
		Methods("OPTIONS", "POST")
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	AccessToken     string                `json:"access_token"`
	OnUserDuplicate model.OnUserDuplicate `json:"on_user_duplicate"`
}

func (p *LoginRequestPayload) SetDefaultValue() {
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

type OAuthLoginInteractionFlow interface {
	LoginWithOAuthProvider(
		clientID string, oauthAuthInfo sso.AuthInfo, codeChallenge string, onUserDuplicate model.OnUserDuplicate,
	) (string, error)

	ExchangeCode(codeHash string, verifier string) (*interactionflows.AuthResult, error)
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
	TxContext     db.TxContext
	Validator     *validation.Validator
	SSOProvider   sso.Provider
	Interactions  OAuthLoginInteractionFlow
	OAuthProvider sso.OAuthProvider
}

func (h LoginHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
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

	var result *interactionflows.AuthResult
	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Handle(req, payload)
		return
	})
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	result.WriteResponse(resp)
}

func (h LoginHandler) Handle(r *http.Request, payload LoginRequestPayload) (*interactionflows.AuthResult, error) {
	if !h.SSOProvider.IsExternalAccessTokenFlowEnabled() {
		return nil, skyerr.NewNotFound("external access token flow is disabled")
	}

	provider, ok := h.OAuthProvider.(sso.ExternalAccessTokenFlowProvider)
	if !ok {
		return nil, skyerr.NewNotFound("unknown provider")
	}

	loginState := sso.LoginState{
		OnUserDuplicate: payload.OnUserDuplicate,
	}

	oauthAuthInfo, err := provider.ExternalAccessTokenGetAuthInfo(sso.NewBearerAccessTokenResp(payload.AccessToken))
	if err != nil {
		return nil, err
	}

	code, err := h.Interactions.LoginWithOAuthProvider(
		coreauth.GetAccessKey(r.Context()).Client.ClientID(),
		oauthAuthInfo,
		"",
		loginState.OnUserDuplicate,
	)
	if err != nil {
		return nil, err
	}

	result, err := h.Interactions.ExchangeCode(code, "")
	if err != nil {
		return nil, err
	}

	return result, nil
}
