package sso

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
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
	h.OAuthProvider = h.ProviderFactory.NewOAuthProvider(h.ProviderID)
	return h
}

// AuthRequestPayload is sso.OAuthAuthorizationResponse
type AuthRequestPayload sso.OAuthAuthorizationResponse

// Validate request payload
func (p AuthRequestPayload) Validate() error {
	if p.Code == "" {
		return errors.New("code is required")
	}

	if p.State == "" {
		return errors.New("state is required")
	}

	if p.Nonce == "" {
		return errors.New("nonce is required")
	}

	return nil
}

// AuthHandler decodes code response and fetch access token from provider.
type AuthHandler struct {
	TxContext                      db.TxContext                `dependency:"TxContext"`
	AuthContext                    coreAuth.ContextGetter      `dependency:"AuthContextGetter"`
	AuthContextSetter              coreAuth.ContextSetter      `dependency:"AuthContextSetter"`
	APIClientConfigurationProvider apiclientconfig.Provider    `dependency:"APIClientConfigurationProvider"`
	OAuthAuthProvider              oauth.Provider              `dependency:"OAuthAuthProvider"`
	IdentityProvider               principal.IdentityProvider  `dependency:"IdentityProvider"`
	AuthInfoStore                  authinfo.Store              `dependency:"AuthInfoStore"`
	AuthnSessionProvider           authnsession.Provider       `dependency:"AuthnSessionProvider"`
	AuthHandlerHTMLProvider        sso.AuthHandlerHTMLProvider `dependency:"AuthHandlerHTMLProvider"`
	ProviderFactory                *sso.OAuthProviderFactory   `dependency:"SSOOAuthProviderFactory"`
	UserProfileStore               userprofile.Store           `dependency:"UserProfileStore"`
	HookProvider                   hook.Provider               `dependency:"HookProvider"`
	WelcomeEmailEnabled            bool                        `dependency:"WelcomeEmailEnabled"`
	TaskQueue                      async.Queue                 `dependency:"AsyncTaskQueue"`
	URLPrefix                      *url.URL                    `dependency:"URLPrefix"`
	SSOProvider                    sso.Provider                `dependency:"SSOProvider"`
	OAuthProvider                  sso.OAuthProvider
	ProviderID                     string
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

// dummy error
type authHandlerError struct{}

func (authHandlerError) Error() string {
	return ""
}

func (h AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hook.WithTx(h.HookProvider, h.TxContext, func() error {
		success := h.Handle(w, r)
		if !success {
			return authHandlerError{}
		}
		return nil
	})
}

func (h AuthHandler) Handle(w http.ResponseWriter, r *http.Request) (success bool) {
	var code string
	var oauthAuthInfo sso.AuthInfo
	success = false

	// We have to return error by directly writing to response at this stage
	// because we do not have valid state.
	if h.OAuthProvider == nil {
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}

	payload, err := h.DecodeRequest(r)
	if err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}

	if err = payload.Validate(); err != nil {
		http.Error(w, "Failed to validate request", http.StatusBadRequest)
		return
	}

	reqPayload := payload.(AuthRequestPayload)

	state, err := h.SSOProvider.DecodeState(reqPayload.State)
	if err != nil {
		http.Error(w, "Failed to decode state", http.StatusBadRequest)
		return
	}

	// Extract API Key from state
	key := h.APIClientConfigurationProvider.GetAccessKeyByClientID(state.APIClientID)
	h.AuthContextSetter.SetAccessKey(key)

	if !h.SSOProvider.IsValidCallbackURL(state.CallbackURL) {
		http.Error(w, "Invalid callback URL", http.StatusBadRequest)
		return
	}

	// From now on, we must return response by respecting CallbackURL and UXMode.
	defer func() {
		success = err == nil
		switch state.UXMode {
		case sso.UXModeWebRedirect, sso.UXModeWebPopup:
			err = h.handleWebAppResponse(w, r, state.UXMode, state.CallbackURL, code, err)
		case sso.UXModeMobileApp:
			err = h.handleMobileAppResponse(w, r, state.CallbackURL, code, err)
		default:
			success = false
			http.Error(w, "Invalid UXMode", http.StatusBadRequest)
		}
	}()

	oauthAuthInfo, err = h.OAuthProvider.GetAuthInfo(
		sso.OAuthAuthorizationResponse(reqPayload),
		*state,
	)
	if err != nil {
		return
	}
	code, err = h.handle(oauthAuthInfo, *state)
	return
}

func (h AuthHandler) handle(oauthAuthInfo sso.AuthInfo, state sso.State) (encodedCode string, err error) {
	respHandler := respHandler{
		AuthnSessionProvider: h.AuthnSessionProvider,
		AuthInfoStore:        h.AuthInfoStore,
		OAuthAuthProvider:    h.OAuthAuthProvider,
		IdentityProvider:     h.IdentityProvider,
		UserProfileStore:     h.UserProfileStore,
		HookProvider:         h.HookProvider,
		WelcomeEmailEnabled:  h.WelcomeEmailEnabled,
		TaskQueue:            h.TaskQueue,
		URLPrefix:            h.URLPrefix,
	}

	var code *sso.SkygearAuthorizationCode
	if state.Action == "login" {
		code, err = respHandler.LoginCode(oauthAuthInfo, state.CodeChallenge, state.LoginState)
	} else {
		code, err = respHandler.LinkCode(oauthAuthInfo, state.CodeChallenge, state.LinkState)
	}
	if err != nil {
		return
	}

	encodedCode, err = h.SSOProvider.EncodeSkygearAuthorizationCode(*code)
	if err != nil {
		return
	}

	return
}

func (h AuthHandler) handleWebAppResponse(
	rw http.ResponseWriter,
	r *http.Request,
	uxMode sso.UXMode,
	callbackURL string,
	code string,
	inputErr error,
) (err error) {
	data := make(map[string]interface{})
	data["result"] = makeJSONResponse(code, inputErr)
	data["callback_url"] = callbackURL

	if uxMode == sso.UXModeWebRedirect {
		resultBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		encodedResult := base64.StdEncoding.EncodeToString(resultBytes)
		v := url.Values{}
		v.Set("x-skygear-result", encodedResult)
		u, err := url.Parse(callbackURL)
		if err != nil {
			return err
		}
		u.RawQuery = v.Encode()
		http.Redirect(rw, r, u.String(), http.StatusFound)
	} else {
		html, err := h.AuthHandlerHTMLProvider.HTML(data)
		if err != nil {
			return err
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(rw, html)
	}
	return
}

func (h AuthHandler) handleMobileAppResponse(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	code string,
	inputErr error,
) error {
	data := make(map[string]interface{})
	data["result"] = makeJSONResponse(code, inputErr)
	data["callback_url"] = callbackURL

	resultBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	encodedResult := base64.StdEncoding.EncodeToString(resultBytes)

	v := url.Values{}
	v.Set("x-skygear-result", encodedResult)
	u, err := url.Parse(callbackURL)
	if err != nil {
		return err
	}
	u.RawQuery = v.Encode()
	http.Redirect(rw, r, u.String(), http.StatusFound)
	return nil
}

func makeJSONResponse(code string, err error) handler.APIResponse {
	if err != nil {
		return handler.APIResponse{Error: err}
	}
	return handler.APIResponse{Result: code}
}
