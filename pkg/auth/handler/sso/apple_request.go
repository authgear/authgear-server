package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// The client generates the code verifier.
// The client sends AppleRequestRequest to the Auth Gear.
// The Auth Gear replies with the state and the nonce.
// The client configures ASAuthorizationAppleIDRequest with the state and the nonce.
// The client performs the authorization with ASAuthorizationController.
// The client receives ASAuthorizationAppleIDCredential from delegate callback.
// The client sends the authorization code, the code verifier, the state, the nonce and the scope to the Auth Gear.
// The Auth Gear replies with AuthResponse.
func AttachAppleRequestHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/login_apple_request", &AppleRequestHandlerFactory{
		Dependency: authDependency,
		Action:     "login",
	}).Methods("OPTIONS", "POST")
	server.Handle("/sso/{provider}/link_apple_request", &AppleRequestHandlerFactory{
		Dependency: authDependency,
		Action:     "link",
	}).Methods("OPTIONS", "POST")
	return server
}

type AppleRequestHandlerFactory struct {
	Dependency auth.DependencyMap
	Action     string
}

func (f *AppleRequestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AppleRequestHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.OAuthProvider = h.ProviderFactory.NewOAuthProvider(h.ProviderID)
	h.Action = f.Action
	return h.RequireAuthz(h, h)
}

/*
	@ID AppleRequestRequest
	@RequestBody
		Generate the state and nonce for ASAuthorizationAppleIDRequest.
		@JSONSchema
*/
const AppleRequestRequestSchema = `
{
	"$id": "#AppleRequestRequest",
	"type": "object",
	"properties": {
		"code_challenge": { "type": "string", "minLength": 1 },
		"on_user_duplicate": {"type": "string", "enum": ["abort", "merge", "create"] }
	},
	"required": ["code_challenge"]
}
`

/*
	@ID AppleRequestResponse
	@Response
		The state and nonce.
		@JSONSchema
*/
const AppleRequestResponseSchema = `
{
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"state": { "type": "string" },
				"nonce": { "type": "string" },
				"hashed_nonce": { "type": "string" }
			}
		}
	}
}
`

type AppleRequestResponse struct {
	State       string `json:"state"`
	Nonce       string `json:"nonce"`
	HashedNonce string `json:"hashed_nonce"`
}

type AppleRequestPayload struct {
	CodeChallenge        string                `json:"code_challenge"`
	MergeRealm           string                `json:"-"`
	OnUserDuplicate      model.OnUserDuplicate `json:"on_user_duplicate"`
	PasswordAuthProvider password.Provider     `json:"-"`
	SSOProvider          sso.Provider          `json:"-"`
}

func (p *AppleRequestPayload) SetDefaultValue() {
	if p.MergeRealm == "" {
		p.MergeRealm = password.DefaultRealm
	}

	if p.OnUserDuplicate == "" {
		p.OnUserDuplicate = model.OnUserDuplicateDefault
	}
}

func (p *AppleRequestPayload) Validate() []validation.ErrorCause {
	if !p.PasswordAuthProvider.IsRealmValid(p.MergeRealm) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/merge_realm",
			Message: "merge_realm is not a valid realm",
		}}
	}
	if !p.SSOProvider.IsAllowedOnUserDuplicate(p.OnUserDuplicate) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/on_user_duplicate",
			Message: "on_user_duplicate is not allowed",
		}}
	}
	return nil
}

type AppleRequestHandler struct {
	TxContext                      db.TxContext              `dependency:"TxContext"`
	Validator                      *validation.Validator     `dependency:"Validator"`
	AuthContext                    coreAuth.ContextGetter    `dependency:"AuthContextGetter"`
	RequireAuthz                   handler.RequireAuthz      `dependency:"RequireAuthz"`
	APIClientConfigurationProvider apiclientconfig.Provider  `dependency:"APIClientConfigurationProvider"`
	ProviderFactory                *sso.OAuthProviderFactory `dependency:"SSOOAuthProviderFactory"`
	PasswordAuthProvider           password.Provider         `dependency:"PasswordAuthProvider"`
	SSOProvider                    sso.Provider              `dependency:"SSOProvider"`
	OAuthProvider                  sso.OAuthProvider
	ProviderID                     string
	Action                         string
}

func (h *AppleRequestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *AppleRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := handler.Transactional(h.TxContext, func() (interface{}, error) {
		return h.Handle(w, r)
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}
	handler.WriteResponse(w, handler.APIResponse{Result: result})
}

func (h *AppleRequestHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	if h.OAuthProvider == nil || h.OAuthProvider.Type() != config.OAuthProviderTypeApple {
		err = skyerr.NewBadRequest("expected OAuth provider to be Apple")
		return
	}

	payload := AppleRequestPayload{}
	payload.PasswordAuthProvider = h.PasswordAuthProvider
	payload.SSOProvider = h.SSOProvider
	err = handler.BindJSONBody(r, w, h.Validator, "#AppleRequestRequest", &payload)
	if err != nil {
		return
	}

	apiClientID, _, _ := h.APIClientConfigurationProvider.Get()

	nonce := sso.GenerateOpenIDConnectNonce()

	state := &sso.State{
		LoginState: sso.LoginState{
			MergeRealm:      payload.MergeRealm,
			OnUserDuplicate: payload.OnUserDuplicate,
		},
		Action:        h.Action,
		APIClientID:   apiClientID,
		CodeChallenge: payload.CodeChallenge,
	}
	state.Nonce = crypto.SHA256String(nonce)
	authInfo, _ := h.AuthContext.AuthInfo()
	if authInfo != nil {
		state.UserID = authInfo.ID
	}

	encodedState, err := h.SSOProvider.EncodeState(*state)
	if err != nil {
		return
	}

	result = AppleRequestResponse{
		State:       encodedState,
		Nonce:       nonce,
		HashedNonce: state.Nonce,
	}

	return
}
