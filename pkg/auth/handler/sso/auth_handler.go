package sso

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachAuthHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/auth_handler", &AuthHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "GET", "POST")
	return server
}

type AuthHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderID)
	return h
}

func (f AuthHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf()
}

// AuthRequestPayload is sso.OAuthAuthorizationResponse
type AuthRequestPayload sso.OAuthAuthorizationResponse

// Validate request payload
func (p AuthRequestPayload) Validate() error {
	if p.Code == "" {
		return skyerr.NewInvalidArgument("code is required", []string{"code"})
	}

	if p.State == "" {
		return skyerr.NewInvalidArgument("state is required", []string{"state"})
	}

	if p.Nonce == "" {
		return skyerr.NewInvalidArgument("nonce is required", []string{"nonce"})
	}

	return nil
}

// AuthHandler decodes code response and fetch access token from provider.
//
// curl http://localhost:3000/sso/<provider>/auth_handler?code=<code>&state=<state>
//
// For ux_mode is 'mobile_app',
// it creates a 302 response, and Location points to:
// myapp://user.skygear.io/sso/{provider}/auth_handler?result=
//
// Fox ux_mode is 'web_redirect',
// it creates a 302 response, and Location points to: sso_callback_url
// and set cookie in the response.
//
// For ux_mode is 'web_popup',
// it will render a html page and set cookie in the response.
//
type AuthHandler struct {
	TxContext               db.TxContext                `dependency:"TxContext"`
	AuthContext             coreAuth.ContextGetter      `dependency:"AuthContextGetter"`
	OAuthAuthProvider       oauth.Provider              `dependency:"OAuthAuthProvider"`
	PasswordAuthProvider    password.Provider           `dependency:"PasswordAuthProvider"`
	IdentityProvider        principal.IdentityProvider  `dependency:"IdentityProvider"`
	AuthInfoStore           authinfo.Store              `dependency:"AuthInfoStore"`
	TokenStore              authtoken.Store             `dependency:"TokenStore"`
	AuthHandlerHTMLProvider sso.AuthHandlerHTMLProvider `dependency:"AuthHandlerHTMLProvider"`
	ProviderFactory         *sso.ProviderFactory        `dependency:"SSOProviderFactory"`
	UserProfileStore        userprofile.Store           `dependency:"UserProfileStore"`
	HookProvider            hook.Provider               `dependency:"HookProvider"`
	OAuthConfiguration      config.OAuthConfiguration   `dependency:"OAuthConfiguration"`
	WelcomeEmailEnabled     bool                        `dependency:"WelcomeEmailEnabled"`
	TaskQueue               async.Queue                 `dependency:"AsyncTaskQueue"`
	Provider                sso.OAuthProvider
	ProviderID              string
}

func (h AuthHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthRequestPayload{}
	err := request.ParseForm()
	if err != nil {
		return nil, err
	}
	payload.Code = request.Form.Get("code")
	payload.Scope = request.Form.Get("scope")
	payload.State = request.Form.Get("state")

	cookie, cookieErr := request.Cookie(coreHttp.CookieNameOpenIDConnectNonce)
	if cookieErr != http.ErrNoCookie {
		payload.Nonce = cookie.Value
	}

	return payload, nil
}

func (h AuthHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var ok interface{}
	var err error
	var oauthAuthInfo sso.AuthInfo

	// We have to return error by directly writing to response at this stage
	// because we do not have valid state.
	if h.Provider == nil {
		http.Error(rw, "Provider is not supported", http.StatusBadRequest)
		return
	}

	payload, err := h.DecodeRequest(r)
	if err != nil {
		http.Error(rw, "Failed to decode request", http.StatusBadRequest)
		return
	}

	if err = payload.Validate(); err != nil {
		http.Error(rw, "Failed to validate request", http.StatusBadRequest)
		return
	}

	if e := h.TxContext.BeginTx(); e != nil {
		http.Error(rw, "Internal Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			h.TxContext.RollbackTx()
		} else {
			h.TxContext.CommitTx()
		}
	}()

	reqPayload := payload.(AuthRequestPayload)

	state, err := h.Provider.DecodeState(reqPayload.State)
	if err != nil {
		http.Error(rw, "Failed to decode state", http.StatusBadRequest)
		return
	}

	if err = h.validateCallbackURL(h.OAuthConfiguration.AllowedCallbackURLs, state.CallbackURL); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// From now on, we must return response by respecting CallbackURL and UXMode.
	defer func() {
		switch state.UXMode {
		case sso.UXModeWebRedirect, sso.UXModeWebPopup:
			_ = h.handleSessionResp(rw, r, state.UXMode, state.CallbackURL, ok, err)
		case sso.UXModeMobileApp:
			_ = h.handleRedirectResp(rw, r, state.CallbackURL, ok, err)
		default:
			http.Error(rw, "Invalid UXMode", http.StatusBadRequest)
		}
	}()

	oauthAuthInfo, err = h.getAuthInfo(reqPayload)
	if err != nil {
		return
	}
	ok, err = h.handle(oauthAuthInfo, *state)
	if err != nil {
		return
	}
}

func (h AuthHandler) getAuthInfo(payload AuthRequestPayload) (oauthAuthInfo sso.AuthInfo, err error) {
	oauthAuthInfo, err = h.Provider.GetAuthInfo(sso.OAuthAuthorizationResponse(payload))
	if err != nil {
		if ssoErr, ok := err.(sso.Error); ok {
			switch ssoErr.Code() {
			case sso.InvalidGrant:
				err = skyerr.NewError(skyerr.InvalidArgument, "Code was already redeemed")
			case sso.InvalidClient:
				err = skyerr.NewError(skyerr.InvalidCredentials, "Unauthorized, please check the app client id and secret")
			default:
				err = skyerr.NewError(skyerr.InvalidCredentials, ssoErr.Error())
			}
		}
	}
	return
}

func (h AuthHandler) handle(oauthAuthInfo sso.AuthInfo, state sso.State) (resp interface{}, err error) {
	respHandler := respHandler{
		TokenStore:           h.TokenStore,
		AuthInfoStore:        h.AuthInfoStore,
		OAuthAuthProvider:    h.OAuthAuthProvider,
		PasswordAuthProvider: h.PasswordAuthProvider,
		IdentityProvider:     h.IdentityProvider,
		UserProfileStore:     h.UserProfileStore,
		HookProvider:         h.HookProvider,
		WelcomeEmailEnabled:  h.WelcomeEmailEnabled,
		TaskQueue:            h.TaskQueue,
	}

	if state.Action == "login" {
		return respHandler.loginActionResp(oauthAuthInfo, state.LoginState)
	}

	return respHandler.linkActionResp(oauthAuthInfo, state.LinkState)
}

func (h AuthHandler) validateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	err = sso.ValidateCallbackURL(allowedCallbackURLs, callbackURL)
	if err != nil {
		err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}
	return
}

func (h AuthHandler) handleSessionResp(rw http.ResponseWriter, r *http.Request, uxMode sso.UXMode, callbackURL string, ok interface{}, inputErr error) (err error) {
	//
	// In JS oauth flow, result send through cookies and handler by js script
	//
	// Session data:
	// sso_callback_url -- callback url for ux_mode == web_redirect
	// sso_result       -- response json
	//
	data := make(map[string]interface{})
	data["result"] = makeJSONResponse(ok, inputErr)
	data["callback_url"] = callbackURL
	msg, err := json.Marshal(data)
	if err != nil {
		return
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	cookie := http.Cookie{
		Name:  coreHttp.CookieNameSSOData,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(rw, &cookie)
	if uxMode == sso.UXModeWebRedirect {
		http.Redirect(rw, r, callbackURL, http.StatusFound)
	} else {
		html, err := h.AuthHandlerHTMLProvider.HTML()
		if err != nil {
			return err
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(rw, html)
	}
	return
}

func (h AuthHandler) handleRedirectResp(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	ok interface{},
	inputErr error,
) error {
	// In mobile app oauth flow, after auth flow complete will redirect
	// client back to the app with custom scheme
	// result will be added to the url by query
	//
	// Example:
	// myapp://user.skygear.io/sso/{provider}/auth_handler?result=
	authRespBytes, err := json.Marshal(makeJSONResponse(ok, inputErr))
	if err != nil {
		return err
	}
	encodedResult := base64.StdEncoding.EncodeToString(authRespBytes)
	v := url.Values{}
	v.Set("result", encodedResult)
	u, err := url.Parse(callbackURL)
	if err != nil {
		return err
	}
	u.RawQuery = v.Encode()
	http.Redirect(rw, r, u.String(), http.StatusFound)
	return nil
}

func makeJSONResponse(ok interface{}, err error) handler.APIResponse {
	if err != nil {
		return handler.APIResponse{
			Err: skyerr.MakeError(err),
		}
	}
	return handler.APIResponse{
		Result: ok,
	}
}
