package sso

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

func AttachAuthHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/auth_handler").
		Handler(auth.MakeHandler(authDependency, newAuthHandler)).
		Methods("OPTIONS", "GET", "POST")
}

type AuthRequestPayload sso.OAuthAuthorizationResponse

// Validate request payload
func (p AuthRequestPayload) Validate() error {
	if p.Code == "" {
		return errors.New("code is required")
	}

	if p.State == "" {
		return errors.New("state is required")
	}

	return nil
}

type AuthHandlerAuthnProvider interface {
	OAuthAuthenticateCode(
		authInfo sso.AuthInfo,
		codeChallenge string,
		loginState sso.LoginState,
	) (*sso.SkygearAuthorizationCode, string, error)

	OAuthLinkCode(
		authInfo sso.AuthInfo,
		codeChallenge string,
		linkState sso.LinkState,
	) (*sso.SkygearAuthorizationCode, string, error)
}

// AuthHandler decodes code response and fetch access token from provider.
type AuthHandler struct {
	TxContext               db.TxContext
	TenantConfiguration     *config.TenantConfiguration
	AuthHandlerHTMLProvider sso.AuthHandlerHTMLProvider
	SSOProvider             sso.Provider
	AuthnProvider           AuthHandlerAuthnProvider
	OAuthProvider           sso.OAuthProvider
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

	return payload, nil
}

// dummy error
type authHandlerError struct{}

func (authHandlerError) Error() string {
	return ""
}

func (h AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db.WithTx(h.TxContext, func() error {
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
	client, _ := model.GetClientConfig(h.TenantConfiguration.AppConfig.Clients, state.APIClientID)

	apiSSOState := AuthAPISSOState(state.Extra)
	if !h.SSOProvider.IsValidCallbackURL(client, apiSSOState.CallbackURL()) {
		http.Error(w, "Invalid callback URL", http.StatusBadRequest)
		return
	}

	// From now on, we must return response by respecting CallbackURL and UXMode.
	defer func() {
		success = err == nil
		switch state.UXMode {
		case sso.UXModeWebRedirect, sso.UXModeWebPopup:
			err = h.handleWebAppResponse(w, r, state.UXMode, apiSSOState.CallbackURL(), code, err)
		case sso.UXModeMobileApp:
			err = h.handleMobileAppResponse(w, r, apiSSOState.CallbackURL(), code, err)
		case sso.UXModeManual:
			err = h.handleManualResponse(w, code, err)
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
	if err != nil {
		return
	}

	return
}

func (h AuthHandler) handle(oauthAuthInfo sso.AuthInfo, state sso.State) (code string, err error) {
	apiSSOState := AuthAPISSOState(state.Extra)
	if state.Action == "login" {
		_, code, err = h.AuthnProvider.OAuthAuthenticateCode(oauthAuthInfo, apiSSOState.CodeChallenge(), state.LoginState)
	} else {
		_, code, err = h.AuthnProvider.OAuthLinkCode(oauthAuthInfo, apiSSOState.CodeChallenge(), state.LinkState)
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

func (h AuthHandler) handleManualResponse(
	w http.ResponseWriter,
	code string,
	inputErr error,
) error {
	resp := makeJSONResponse(code, inputErr)
	handler.WriteResponse(w, resp)
	return nil
}

func makeJSONResponse(code string, err error) handler.APIResponse {
	if err != nil {
		return handler.APIResponse{Error: err}
	}
	return handler.APIResponse{Result: code}
}
