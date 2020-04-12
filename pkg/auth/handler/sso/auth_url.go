package sso

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type ssoAction string

const (
	ssoActionLogin ssoAction = "login"
	ssoActionLink  ssoAction = "link"
)

func AttachAuthURLHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/login_auth_url").
		Handler(pkg.MakeHandler(authDependency, newLoginAuthURLHandler)).
		Methods("OPTIONS", "POST")

	router.NewRoute().
		Path("/sso/{provider}/link_auth_url").
		Handler(pkg.MakeHandler(authDependency, newLinkAuthURLHandler)).
		Methods("OPTIONS", "POST")
}

// nolint: deadcode
/*
	@ID SSOCallbackURL
	@Parameter callback_url query
		Callback URL after SSO flow
		@JSONSchema
			{ "type": "string" }
*/
type ssoCallbackURL string

// nolint: deadcode
/*
	@ID SSOUXMode
	@Parameter ux_mode query
		UX mode of SSO flow
		@JSONSchema
			{ "type": "string" }
*/
type ssoUXMode string

// nolint: deadcode
/*
	@ID SSOOnUserDuplicate
	@Parameter on_user_duplicate query
		Behavior when duplicated user is detected
		@JSONSchema
			{ "type": "string" }
*/
type ssoOnUserDuplicate string

/*
	@ID AuthURLRequest
	@RequestBody
		Describe desired behavior and UX of SSO flow.
		@JSONSchema
*/
const AuthURLRequestSchema = `
{
	"$id": "#AuthURLRequest",
	"type": "object",
	"properties": {
		"code_challenge": { "type": "string", "minLength": 1 },
		"callback_url": { "type": "string", "format": "uri" },
		"ux_mode": { "type": "string", "enum": ["web_redirect", "web_popup", "mobile_app", "manual"] },
		"on_user_duplicate": {"type": "string", "enum": ["abort", "merge", "create"] }
	},
	"required": ["code_challenge", "callback_url", "ux_mode"]
}
`

/*
	@ID AuthURLResponse
	@Response
		SSO initiation URL.
		@JSONSchema
		@JSONExample Success - Return SSO URL
		{
			"result": "https://myapp.skygearapis.com/_auth/sso/provider/auth_handler"
		}
*/
const AuthURLResponseSchema = `
{
	"type": "object",
	"properties": {
		"result": { "type": "string" }
	}
}
`

// AuthURLRequestPayload login handler request payload
type AuthURLRequestPayload struct {
	CodeChallenge   string                `json:"code_challenge"`
	CallbackURL     string                `json:"callback_url"`
	UXMode          sso.UXMode            `json:"ux_mode"`
	MergeRealm      string                `json:"-"`
	OnUserDuplicate model.OnUserDuplicate `json:"on_user_duplicate"`

	Client               config.OAuthClientConfiguration           `json:"-"`
	PasswordAuthProvider password.Provider                         `json:"-"`
	SSOProvider          sso.Provider                              `json:"-"`
	ConflictConfig       *config.AuthAPIOAuthConflictConfiguration `json:"-"`
}

func (p *AuthURLRequestPayload) SetDefaultValue() {
	if p.MergeRealm == "" {
		p.MergeRealm = password.DefaultRealm
	}

	if p.OnUserDuplicate == "" {
		p.OnUserDuplicate = model.OnUserDuplicateDefault
	}
}

func (p *AuthURLRequestPayload) Validate() []validation.ErrorCause {
	if !p.PasswordAuthProvider.IsRealmValid(p.MergeRealm) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/merge_realm",
			Message: "merge_realm is not a valid realm",
		}}
	}

	if !model.IsAllowedOnUserDuplicate(
		p.ConflictConfig.AllowAutoMergeUser,
		p.ConflictConfig.AllowCreateNewUser,
		p.OnUserDuplicate,
	) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/on_user_duplicate",
			Message: "on_user_duplicate is not allowed",
		}}
	}

	if !p.SSOProvider.IsValidCallbackURL(p.Client, p.CallbackURL) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/callback_url",
			Message: "callback_url is not allowed",
		}}
	}

	return nil
}

/*
	@Operation POST /sso/{provider_id}/login_auth_url - Get login SSO flow url of provider
		Returns SSO auth URL. Client should redirect user agent to this URL to
		initiate SSO login flow.

		@Tag SSO

		@Parameter {SSOProviderID}
		@RequestBody {AuthURLRequest}
		@Response 200 {AuthURLResponse}

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}

	@Operation POST /sso/{provider_id}/link_auth_url - Get link SSO link url of provider
		Returns SSO auth URL. Client should redirect user agent to this URL to
		initiate SSO link flow.

		@Tag SSO

		@Parameter {SSOProviderID}
		@RequestBody {AuthURLRequest}
		@Response 200 {AuthURLResponse}

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthURLHandler struct {
	TxContext                  db.TxContext
	Validator                  *validation.Validator
	RequireAuthz               handler.RequireAuthz
	PasswordAuthProvider       password.Provider
	SSOProvider                sso.Provider
	OAuthConflictConfiguration *config.AuthAPIOAuthConflictConfiguration
	OAuthProvider              sso.OAuthProvider
	Action                     ssoAction
}

func (h *AuthURLHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return policy.AllOf(
		authz.AuthAPIRequireClient,
		coreauthz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *AuthURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := handler.Transactional(h.TxContext, func() (interface{}, error) {
		return h.Handle(w, r)
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}
	handler.WriteResponse(w, handler.APIResponse{Result: result})
}

func (h *AuthURLHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	if h.OAuthProvider == nil {
		err = skyerr.NewNotFound("unknown provider")
		return
	}

	vars := mux.Vars(r)
	providerID := vars["provider"]

	payload := AuthURLRequestPayload{}
	payload.Client = coreauth.GetAccessKey(r.Context()).Client
	payload.PasswordAuthProvider = h.PasswordAuthProvider
	payload.SSOProvider = h.SSOProvider
	payload.ConflictConfig = h.OAuthConflictConfiguration
	err = handler.BindJSONBody(r, w, h.Validator, "#AuthURLRequest", &payload)
	if err != nil {
		return
	}

	apiClientID := payload.Client.ClientID()

	// The information in the state are mostly from the client.
	// APIClientID is derived from the API key used by the client.
	// UserID is derived from the access token.
	// The state is then signed but no encrypted and returned to the client.
	state := sso.State{
		LoginState: sso.LoginState{
			MergeRealm:      payload.MergeRealm,
			OnUserDuplicate: payload.OnUserDuplicate,
		},
		OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
			CallbackURL: payload.CallbackURL,
			UXMode:      payload.UXMode,
		},
		Action:        string(h.Action),
		APIClientID:   apiClientID,
		CodeChallenge: payload.CodeChallenge,
	}
	session := auth.GetSession(r.Context())
	if session != nil {
		state.UserID = session.AuthnAttrs().UserID
	}

	encodedState, err := h.SSOProvider.EncodeState(state)
	if err != nil {
		return
	}

	q := url.Values{}
	q.Set("state", encodedState)

	u := &url.URL{
		Host:     coreHttp.GetHost(r),
		Scheme:   coreHttp.GetProto(r),
		Path:     fmt.Sprintf("/_auth/sso/%s/auth_redirect", providerID),
		RawQuery: q.Encode(),
	}

	result = u.String()
	return
}
