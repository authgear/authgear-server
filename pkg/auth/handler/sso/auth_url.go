package sso

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
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
	return h
}

func (f AuthURLHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.Everybody{Allow: true}
}

// AuthURLRequestPayload login handler request payload
type AuthURLRequestPayload struct {
	CallbackURL     string              `json:"callback_url"`
	UXMode          sso.UXMode          `json:"ux_mode"`
	MergeRealm      string              `json:"merge_realm"`
	OnUserDuplicate sso.OnUserDuplicate `json:"on_user_duplicate"`
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

	if !sso.IsValidOnUserDuplicate(p.OnUserDuplicate) {
		err = skyerr.NewInvalidArgument("Invalid OnUserDuplicate", []string{"on_user_duplicate"})
		return
	}
	return
}

// AuthURLHandler returns the SSO auth url by provider.
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/<provider>/login_auth_url \
// <<EOF
// {
//     callback_url: <url>,
//     ux_mode: <ux_mode>
// }
// EOF
//
// {
//     "result": "<auth_url>"
// }
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/<provider>/link_auth_url \
// <<EOF
// {
//     callback_url: <url>,
//     ux_mode: <ux_mode>
// }
// EOF
//
// {
//     "result": "<auth_url>"
// }
// The handler also supports GET method. If you are experimenting
// with an OpenID Connect provider, you should construct an URL
// and visit it in a browser. In this way, nonce is set in the session
// cookie and automatically redirected to the provider authorization URL.
type AuthURLHandler struct {
	TxContext            db.TxContext              `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter    `dependency:"AuthContextGetter"`
	ProviderFactory      *sso.ProviderFactory      `dependency:"SSOProviderFactory"`
	PasswordAuthProvider password.Provider         `dependency:"PasswordAuthProvider"`
	OAuthConfiguration   config.OAuthConfiguration `dependency:"OAuthConfiguration"`
	Provider             sso.OAuthProvider
	ProviderID           string
	Action               string
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
		err = json.NewDecoder(r.Body).Decode(&payload)
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
		payload.OnUserDuplicate = sso.OnUserDuplicate(r.Form.Get("on_user_duplicate"))
	}

	if payload.MergeRealm == "" {
		payload.MergeRealm = password.DefaultRealm
	}

	if payload.OnUserDuplicate == "" {
		payload.OnUserDuplicate = sso.OnUserDuplicateDefault
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

	if !sso.IsAllowedOnUserDuplicate(
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
		},
		Nonce: cookie.Value,
	}
	if h.AuthContext.AuthInfo() != nil {
		params.State.UserID = h.AuthContext.AuthInfo().ID
	}
	url, err := h.Provider.GetAuthURL(params)
	if err != nil {
		return
	}
	result = url
	return
}
