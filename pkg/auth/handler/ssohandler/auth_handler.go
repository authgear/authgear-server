package ssohandler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
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
	// since auth_hander need create different responses depends on ux_mode,
	// so here has a APIHandler to handle those different situation.
	return h.Handler()
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
	Provider                sso.Provider                `dependency:"SSOProvider"`
	OAuthAuthProvider       oauth.Provider              `dependency:"OAuthAuthProvider"`
	AuthInfoStore           authinfo.Store              `dependency:"AuthInfoStore"`
	RoleStore               role.Store                  `dependency:"RoleStore"`
	TokenStore              authtoken.Store             `dependency:"TokenStore"`
	AuthHandlerHTMLProvider sso.AuthHandlerHTMLProvider `dependency:"AuthHandlerHTMLProvider"`
	ProviderName            string
}

func (h AuthHandler) WithTx() bool {
	return true
}

func (h AuthHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthRequestPayload{}
	q := request.URL.Query()
	payload.Code = q.Get("code")
	payload.Scope = strings.Split(q.Get("scope"), " ")
	payload.EncodedState = q.Get("state")

	return payload, nil
}

func (h AuthHandler) Handler() http.Handler {
	// reference from APIHandlerToHandler
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var respErr skyerr.Error

		if h.Provider == nil {
			respErr = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderName})
			return
		}

		defer func() {
			if respErr != nil {
				response := handler.APIResponse{
					Err: respErr,
				}
				handler.WriteResponse(rw, response)
			}
		}()

		payload, err := h.DecodeRequest(r)
		if err != nil {
			respErr = skyerr.MakeError(err)
			return
		}

		if err := payload.Validate(); err != nil {
			respErr = skyerr.MakeError(err)
			return
		}

		if h.WithTx() {
			// assume txContext != nil if apiHandler.WithTx() is true
			if err := h.TxContext.BeginTx(); err != nil {
				panic(err)
			}

			defer func() {
				if h.TxContext.HasTx() {
					h.TxContext.RollbackTx()
				}
			}()
		}

		reqPayload := payload.(AuthRequestPayload)
		oauthAuthInfo, err := h.getAuthInfo(reqPayload)
		if err != nil {
			respErr = skyerr.MakeError(err)
			return
		}
		authResp, err := h.getAuthResp(oauthAuthInfo)
		if h.TxContext != nil {
			h.TxContext.CommitTx()
		}

		// handle callback url by ux_mode
		tConfig := config.GetTenantConfig(r)
		SSOSetting := tConfig.SSOSetting
		UXMode := oauthAuthInfo.State.UXMode
		callbackURL := oauthAuthInfo.State.CallbackURL
		allowedCallbackURLs := SSOSetting.AllowedCallbackURLs
		err = h.validateCallbackURL(allowedCallbackURLs, callbackURL)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		// handle authResp by UXMode
		switch oauthAuthInfo.State.UXMode {
		case sso.WebRedirect.String(), sso.WebPopup.String():
			h.handleSessionResp(rw, r, UXMode, callbackURL, authResp)
		case sso.IOS.String(), sso.Android.String():
			h.handleRedirectResp(rw, r, UXMode, callbackURL, authResp)
		}
	})
}

func (h AuthHandler) getAuthInfo(payload AuthRequestPayload) (oauthAuthInfo sso.AuthInfo, err error) {
	oauthAuthInfo, err = h.Provider.GetAuthInfo(payload.Code, payload.Scope, payload.EncodedState)
	if err != nil {
		if ssoErr, ok := err.(sso.Error); ok {
			switch ssoErr.Code() {
			case sso.InvalidGrant:
				err = skyerr.NewError(skyerr.InvalidArgument, "Code was already redeemed")
			case sso.InvalidClient:
				err = skyerr.NewError(skyerr.InvalidCredentials, "auth_data or password incorrect")
			default:
				err = skyerr.NewError(skyerr.InvalidCredentials, ssoErr.Error())
			}
		} else {
			return
		}
	}
	return
}

func (h AuthHandler) getAuthResp(oauthAuthInfo sso.AuthInfo) (resp response.AuthResponse, err error) {
	if oauthAuthInfo.State.Action == "login" {
		var info authinfo.AuthInfo
		err = h.handleLogin(&info, oauthAuthInfo)
		if err != nil {
			return
		}

		// Create auth token
		var token authtoken.Token
		token, err = h.TokenStore.NewToken(info.ID)
		if err != nil {
			panic(err)
		}
		if err = h.TokenStore.Put(&token); err != nil {
			panic(err)
		}

		// TODO: convert oauthAuthInfo.UserProfile to userprofile.UserProfile
		var userProfile userprofile.UserProfile
		resp = response.NewAuthResponse(info, userProfile, token.AccessToken)

		// Populate the activity time to user
		now := timeNow()
		info.LastSeenAt = &now
		if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
			err = skyerr.MakeError(err)
			return
		}
	} else {
		// TODO: handle link action
	}

	return
}

func (h AuthHandler) handleLogin(info *authinfo.AuthInfo, oauthAuthInfo sso.AuthInfo) (err error) {
	createNewUser := false
	now := timeNow()

	principal, err := h.OAuthAuthProvider.GetPrincipalByUserID(oauthAuthInfo.ProviderName, oauthAuthInfo.UserID)
	if err != nil {
		if err != skydb.ErrUserNotFound {
			return
		}
		err = nil
		createNewUser = true
	}

	if createNewUser {
		// TODO: check auto connect user flow
		// 1. find existed user
		// 2. link user
		// 3. login user

		// if there is no existed user
		// signup a new user
		*info = authinfo.NewAuthInfo()
		info.LastLoginAt = &now

		// Get default roles
		defaultRoles, e := h.RoleStore.GetDefaultRoles()
		if e != nil {
			err = skyerr.NewError(skyerr.InternalQueryInvalid, "unable to query default roles")
			return
		}

		// Assign default roles
		info.Roles = defaultRoles

		// Create AuthInfo
		if e = h.AuthInfoStore.CreateAuth(info); e != nil {
			if e == skydb.ErrUserDuplicated {
				err = skyerr.NewError(skyerr.Duplicated, "user duplicated")
				return
			}
			// TODO:
			// return proper error
			err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save auth info")
			return
		}

		principal := oauth.NewPrincipal()
		principal.UserID = info.ID
		principal.ProviderName = oauthAuthInfo.ProviderName
		principal.ProviderUserID = oauthAuthInfo.UserID
		principal.AccessTokenResp = oauthAuthInfo.AccessTokenResp
		principal.UserProfile = oauthAuthInfo.UserProfile
		principal.CreatedAt = &now
		principal.UpdatedAt = &now
		err = h.OAuthAuthProvider.CreatePrincipal(principal)
	} else {
		principal.AccessTokenResp = oauthAuthInfo.AccessTokenResp
		principal.UserProfile = oauthAuthInfo.UserProfile
		principal.UpdatedAt = &now

		if err = h.OAuthAuthProvider.UpdatePrincipal(principal); err != nil {
			err = skyerr.MakeError(err)
			return
		}

		if e := h.AuthInfoStore.GetAuth(principal.UserID, info); e != nil {
			if err == skydb.ErrUserNotFound {
				err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
				return
			}
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
	}
	return
}

func (h AuthHandler) validateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	if callbackURL == "" {
		err = errors.New("Missing callback url")
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
			err = errors.New("The callback url is not whitelisted in the social login setting")
		}
	}

	return
}

func (h AuthHandler) handleSessionResp(rw http.ResponseWriter, r *http.Request, UXMode string, callbackURL string, authResp response.AuthResponse) {
	/*
	   In JS oauth flow, result send through cookies and handler by js script

	   Session data:
	   sso_callback_url -- callback url for ux_mode == web_redirect
	   sso_result       -- response json
	*/
	data := make(map[string]interface{})
	data["result"] = authResp
	data["callback_url"] = callbackURL
	msg, _ := json.Marshal(data)
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	cookie := http.Cookie{
		Name:  "sso_data",
		Value: encoded,
	}
	if UXMode == sso.WebRedirect.String() {
		http.SetCookie(rw, &cookie)
		http.Redirect(rw, r, callbackURL, http.StatusFound)
	} else {
		html, _ := h.AuthHandlerHTMLProvider.HTML()
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(rw, html)
	}
}

func (h AuthHandler) handleRedirectResp(rw http.ResponseWriter, r *http.Request, UXMode string, callbackURL string, authResp response.AuthResponse) {
	/*
	   In ios and android oauth flow, after auth flow complete will redirect
	   client back to the app with custom scheme
	   result will be added to the url by query

	   Example:
	   myapp://user.skygear.io/sso/{provider}/auth_handler?result=
	*/
	authRespBytes, _ := json.Marshal(authResp)
	encodedResult := base64.StdEncoding.EncodeToString(authRespBytes)
	v := url.Values{}
	v.Set("result", encodedResult)
	u, _ := url.Parse(callbackURL)
	u.RawQuery = v.Encode()
	http.Redirect(rw, r, u.String(), http.StatusFound)
}
