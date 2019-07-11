package sso

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachAuthHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/auth_handler", &AuthHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "GET")
	return server
}

type AuthHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderName = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderName)
	h.SSOSetting = h.ProviderFactory.Setting()
	return h
}

func (f AuthHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf()
}

// AuthRequestPayload login handler request payload
type AuthRequestPayload struct {
	Code         string
	Scope        sso.Scope
	EncodedState string
}

// Validate request payload
func (p AuthRequestPayload) Validate() error {
	if p.Code == "" {
		return skyerr.NewInvalidArgument("Authorization Code is required", []string{"code"})
	}

	if p.EncodedState == "" {
		return skyerr.NewInvalidArgument("EncodedState is required", []string{"state"})
	}

	return nil
}

// AuthHandler decodes code response and fetch access token from provider.
//
// curl http://localhost:3000/sso/<provider>/auth_handler?code=<code>&state=<state>
//
// For ux_mode is 'ios' or 'android',
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
	Provider                sso.Provider
	SSOSetting              sso.Setting
	ProviderName            string
}

func (h AuthHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthRequestPayload{}
	q := request.URL.Query()
	payload.Code = q.Get("code")
	payload.Scope = strings.Split(q.Get("scope"), " ")
	payload.EncodedState = q.Get("state")

	return payload, nil
}

func (h AuthHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var err error
	var oauthAuthInfo sso.AuthInfo
	var resp interface{}

	if h.Provider == nil {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderName})
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	payload, err := h.DecodeRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if err = payload.Validate(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.TxContext.BeginTx(); err != nil {
		panic(err)
	}

	defer func() {
		if err != nil {
			h.TxContext.RollbackTx()
		} else {
			h.TxContext.CommitTx()
		}
	}()

	reqPayload := payload.(AuthRequestPayload)
	oauthAuthInfo, err = h.getAuthInfo(reqPayload)
	c := authHandlerRespContext{
		callbackURL: oauthAuthInfo.State.CallbackURL,
		UXMode:      oauthAuthInfo.State.UXMode,
		err:         err,
	}
	if err != nil {
		// send back resp depends on different uxmode
		err = h.sendResp(rw, r, c)
		return
	}

	// get resp depends on different action
	resp, err = h.getResp(oauthAuthInfo)
	c.succ = resp
	c.err = err

	// send back resp depends on different uxmode
	err = h.sendResp(rw, r, c)
}

func (h AuthHandler) getAuthInfo(payload AuthRequestPayload) (oauthAuthInfo sso.AuthInfo, err error) {
	oauthAuthInfo, err = h.Provider.GetAuthInfo(payload.Code, payload.Scope, payload.EncodedState)
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

func (h AuthHandler) getResp(oauthAuthInfo sso.AuthInfo) (resp interface{}, err error) {
	respHandler := respHandler{
		TokenStore:           h.TokenStore,
		AuthInfoStore:        h.AuthInfoStore,
		OAuthAuthProvider:    h.OAuthAuthProvider,
		PasswordAuthProvider: h.PasswordAuthProvider,
		IdentityProvider:     h.IdentityProvider,
		UserProfileStore:     h.UserProfileStore,
		UserID:               oauthAuthInfo.State.UserID,
		Settings:             h.SSOSetting,
	}

	if oauthAuthInfo.State.Action == "login" {
		return respHandler.loginActionResp(oauthAuthInfo)
	}

	return respHandler.linkActionResp(oauthAuthInfo)
}

func (h AuthHandler) validateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	if callbackURL == "" {
		err = skyerr.NewError(skyerr.BadRequest, "Missing callback url")
		return
	}
	if len(allowedCallbackURLs) != 0 {
		found := false
		lowerCallbackURL := strings.ToLower(callbackURL)
		for _, v := range allowedCallbackURLs {
			lowerAllowed := strings.ToLower(v)
			if strings.HasPrefix(lowerCallbackURL, lowerAllowed) {
				found = true
				break
			}
		}

		if !found {
			err = skyerr.NewError(skyerr.BadRequest, "The callback url is not whitelisted in the social login setting")
		}
	}

	return
}

func (h AuthHandler) handleSessionResp(rw http.ResponseWriter, r *http.Request, UXMode string, callbackURL string, resp interface{}) (err error) {
	/*
	   In JS oauth flow, result send through cookies and handler by js script

	   Session data:
	   sso_callback_url -- callback url for ux_mode == web_redirect
	   sso_result       -- response json
	*/
	data := make(map[string]interface{})
	data["result"] = resp
	data["callback_url"] = callbackURL
	msg, err := json.Marshal(data)
	if err != nil {
		return
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	cookie := http.Cookie{
		Name:  "sso_data",
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(rw, &cookie)
	if UXMode == sso.WebRedirect.String() {
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

func (h AuthHandler) handleRedirectResp(rw http.ResponseWriter, r *http.Request, UXMode string, callbackURL string, resp interface{}) (err error) {
	/*
	   In ios and android oauth flow, after auth flow complete will redirect
	   client back to the app with custom scheme
	   result will be added to the url by query

	   Example:
	   myapp://user.skygear.io/sso/{provider}/auth_handler?result=
	*/
	authRespBytes, err := json.Marshal(resp)
	if err != nil {
		return
	}
	encodedResult := base64.StdEncoding.EncodeToString(authRespBytes)
	v := url.Values{}
	v.Set("result", encodedResult)
	u, err := url.Parse(callbackURL)
	if err != nil {
		return
	}
	u.RawQuery = v.Encode()
	http.Redirect(rw, r, u.String(), http.StatusFound)
	return
}

type authHandlerRespContext struct {
	callbackURL string
	UXMode      string
	succ        interface{}
	err         error
}

func (c authHandlerRespContext) generateResp() interface{} {
	// Redirect the result (both success and error) back to the app,
	// so that user can go back to the original app.
	if c.err != nil {
		return handler.APIResponse{
			Err: skyerr.MakeError(c.err),
		}
	}

	// re-wrap resp in result attribute
	return handler.APIResponse{
		Result: c.succ,
	}
}

func (h AuthHandler) sendResp(rw http.ResponseWriter, r *http.Request, c authHandlerRespContext) (err error) {
	if err = h.validateCallbackURL(h.SSOSetting.AllowedCallbackURLs, c.callbackURL); err != nil {
		// there is no callback url for redirect, send 400 bad request instead
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// handle authResp by UXMode
	type authRespHandlerFunc func(rw http.ResponseWriter, r *http.Request, UXMode string, callbackURL string, resp interface{}) (err error)
	var authRespHandler authRespHandlerFunc
	switch c.UXMode {
	case sso.WebRedirect.String(), sso.WebPopup.String():
		authRespHandler = h.handleSessionResp
	case sso.IOS.String(), sso.Android.String():
		authRespHandler = h.handleRedirectResp
	}

	if authRespHandler == nil {
		err = skyerr.NewInvalidArgument("UXMode is not supported", []string{"UXMode"})
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if err = authRespHandler(rw, r, c.UXMode, c.callbackURL, c.generateResp()); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	return
}
