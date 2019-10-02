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
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachAuthURLHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/login_auth_url", &AuthURLHandlerFactory{
		Dependency: authDependency,
		Action:     "login",
	}).Methods("OPTIONS", "POST", "GET")
	server.Handle("/sso/{provider}/link_auth_url", &AuthURLHandlerFactory{
		Dependency: authDependency,
		Action:     "link",
	}).Methods("OPTIONS", "POST", "GET")
	return server
}

type AuthURLHandlerFactory struct {
	Dependency auth.DependencyMap
	Action     string
}

func (f AuthURLHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthURLHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderID)
	h.Action = f.Action
	return h.RequireAuthz(h, h)
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
	@ID SSOMergeRealm
	@Parameter merge_realm query
		Realm to merge when duplicated user is detected
		@JSONSchema
			{ "type": "string" }
*/
type ssoMergeRealm string

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
	"type": "object",
	"properties": {
		"callback_url": { "type": "string" },
		"ux_mode": { "type": "string" },
		"merge_realm": { "type": "string" },
		"on_user_duplicate": { "type": "string" }
	}
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
	CallbackURL     string                `json:"callback_url"`
	UXMode          sso.UXMode            `json:"ux_mode"`
	MergeRealm      string                `json:"merge_realm"`
	OnUserDuplicate model.OnUserDuplicate `json:"on_user_duplicate"`
}

func (p AuthURLRequestPayload) Validate() (err error) {
	if p.CallbackURL == "" {
		err = skyerr.NewInvalidArgument("Callback url is required", []string{"callback_url"})
		return
	}
	if !sso.IsValidUXMode(p.UXMode) {
		err = skyerr.NewInvalidArgument("Invalid UX mode", []string{"ux_mode"})
		return
	}

	if !model.IsValidOnUserDuplicateForSSO(p.OnUserDuplicate) {
		err = skyerr.NewInvalidArgument("Invalid OnUserDuplicate", []string{"on_user_duplicate"})
		return
	}
	return
}

/*
	@Operation POST /sso/{provider_id}/login_auth_url - Get login SSO flow url of provider
		Returns SSO auth URL. Client should redirect user agent to this URL to
		initiate SSO login flow.

		If you are experimenting with an OpenID Connect provider, you should
		use GET method instead visit it in a browser. In this way, nonce is set
		in the session cookie and automatically redirected to the provider
		authorization URL.

		@Tag SSO

		@Parameter {SSOProviderID}
		@RequestBody {AuthURLRequest}
		@Response 200 {AuthURLResponse}

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}

	@Operation GET /sso/{provider_id}/login_auth_url - Begin SSO login flow with provider
		Redirect user to SSO login flow.

		@Tag SSO

		@Parameter {SSOProviderID}
		@Parameter {SSOCallbackURL}
		@Parameter {SSOUXMode}
		@Parameter {SSOMergeRealm}
		@Parameter {SSOOnUserDuplicate}
		@Response 302
			Redirect to SSO login flow

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}

	@Operation POST /sso/{provider_id}/link_auth_url - Get link SSO link url of provider
		Returns SSO auth URL. Client should redirect user agent to this URL to
		initiate SSO link flow.

		If you are experimenting with an OpenID Connect provider, you should
		use GET method instead visit it in a browser. In this way, nonce is set
		in the session cookie and automatically redirected to the provider
		authorization URL.

		@Tag SSO

		@Parameter {SSOProviderID}
		@RequestBody {AuthURLRequest}
		@Response 200 {AuthURLResponse}

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}

	@Operation GET /sso/{provider_id}/link_auth_url - Begin SSO link flow with provider
		Redirect user to SSO link flow.

		@Tag SSO

		@Parameter {SSOProviderID}
		@Parameter {SSOCallbackURL}
		@Parameter {SSOUXMode}
		@Parameter {SSOMergeRealm}
		@Parameter {SSOOnUserDuplicate}
		@Response 302
			Redirect to SSO link flow

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthURLHandler struct {
	TxContext                      db.TxContext              `dependency:"TxContext"`
	AuthContext                    coreAuth.ContextGetter    `dependency:"AuthContextGetter"`
	RequireAuthz                   handler.RequireAuthz      `dependency:"RequireAuthz"`
	APIClientConfigurationProvider apiclientconfig.Provider  `dependency:"APIClientConfigurationProvider"`
	ProviderFactory                *sso.ProviderFactory      `dependency:"SSOProviderFactory"`
	PasswordAuthProvider           password.Provider         `dependency:"PasswordAuthProvider"`
	OAuthConfiguration             config.OAuthConfiguration `dependency:"OAuthConfiguration"`
	Provider                       sso.OAuthProvider
	ProviderID                     string
	Action                         string
}

func (h *AuthURLHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *AuthURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := handler.Transactional(h.TxContext, func() (interface{}, error) {
		return h.Handle(w, r)
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Err: skyerr.MakeError(err)})
		return
	}
	if r.Method == http.MethodPost {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
		return
	}
	http.Redirect(w, r, result.(string), http.StatusFound)
}

func (h *AuthURLHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	payload := AuthURLRequestPayload{}
	if r.Method == http.MethodPost {
		err = handler.DecodeJSONBody(r, w, &payload)
		if err != nil {
			return
		}
	} else {
		err = r.ParseForm()
		if err != nil {
			return
		}
		payload.CallbackURL = r.Form.Get("callback_url")
		payload.UXMode = sso.UXMode(r.Form.Get("ux_mode"))
		payload.MergeRealm = r.Form.Get("merge_realm")
		payload.OnUserDuplicate = model.OnUserDuplicate(r.Form.Get("on_user_duplicate"))
	}

	if payload.MergeRealm == "" {
		payload.MergeRealm = password.DefaultRealm
	}

	if payload.OnUserDuplicate == "" {
		payload.OnUserDuplicate = model.OnUserDuplicateDefault
	}

	err = payload.Validate()
	if err != nil {
		return
	}

	if h.Provider == nil {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderID})
		return
	}

	if !h.PasswordAuthProvider.IsRealmValid(payload.MergeRealm) {
		err = skyerr.NewInvalidArgument("Invalid MergeRealm", []string{payload.MergeRealm})
		return
	}

	if !model.IsAllowedOnUserDuplicate(
		h.OAuthConfiguration.OnUserDuplicateAllowMerge,
		h.OAuthConfiguration.OnUserDuplicateAllowCreate,
		payload.OnUserDuplicate,
	) {
		err = skyerr.NewInvalidArgument("Disallowed OnUserDuplicate", []string{string(payload.OnUserDuplicate)})
		return
	}

	if e := sso.ValidateCallbackURL(h.OAuthConfiguration.AllowedCallbackURLs, payload.CallbackURL); e != nil {
		err = skyerr.NewInvalidArgument(e.Error(), []string{string(payload.CallbackURL)})
		return
	}

	// Always generate a new nonce to ensure it is unpredictable.
	// The developer is expected to call auth_url just before they need to perform the flow.
	// If they call auth_url multiple times ahead of time,
	// only the last auth URL is valid because the nonce of the previous auth URLs are all overwritten.
	nonce := sso.GenerateOpenIDConnectNonce()
	cookie := &http.Cookie{
		Name:     coreHttp.CookieNameOpenIDConnectNonce,
		Value:    nonce,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	apiClientID, _, _ := h.APIClientConfigurationProvider.Get()

	params := sso.GetURLParams{
		State: sso.State{
			LoginState: sso.LoginState{
				MergeRealm:      payload.MergeRealm,
				OnUserDuplicate: payload.OnUserDuplicate,
			},
			OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
				CallbackURL: payload.CallbackURL,
				UXMode:      payload.UXMode,
				Action:      h.Action,
			},
			Nonce:       crypto.SHA256String(nonce),
			APIClientID: apiClientID,
		},
	}
	authInfo, _ := h.AuthContext.AuthInfo()
	if authInfo != nil {
		params.State.UserID = authInfo.ID
	}
	url, err := h.Provider.GetAuthURL(params)
	if err != nil {
		return
	}
	result = url
	return
}
