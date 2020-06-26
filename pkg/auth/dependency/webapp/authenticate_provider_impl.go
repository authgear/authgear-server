package webapp

import (
	"crypto/subtle"
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/httputil"
	"github.com/skygeario/skygear-server/pkg/validation"
)

type InteractionFlow interface {
	LoginWithLoginID(loginID string) (*interactionflows.WebAppResult, error)
	SignupWithLoginID(loginIDKey, loginID string) (*interactionflows.WebAppResult, error)
	PromoteWithLoginID(loginIDKey, loginID string, userID string) (*interactionflows.WebAppResult, error)
	EnterSecret(token string, secret string) (*interactionflows.WebAppResult, error)
	TriggerOOBOTP(token string) (*interactionflows.WebAppResult, error)
	LoginWithOAuthProvider(oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
	LinkWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
	UnlinkWithOAuthProvider(userID string, providerConfig *config.OAuthSSOProviderConfig) (*interactionflows.WebAppResult, error)
	PromoteWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
	AddLoginID(userID string, loginID loginid.LoginID) (*interactionflows.WebAppResult, error)
	UpdateLoginID(userID string, oldLoginID loginid.LoginID, newLoginID loginid.LoginID) (*interactionflows.WebAppResult, error)
	RemoveLoginID(userID string, loginID loginid.LoginID) (*interactionflows.WebAppResult, error)
}

type SSOStateCodec interface {
	EncodeState(state sso.State) (string, error)
	DecodeState(encodedState string) (*sso.State, error)
}

type AuthenticateProviderImpl struct {
	ServerConfig         *config.ServerConfig
	SSOOAuthConfig       *config.OAuthSSOConfig
	ValidateProvider     ValidateProvider
	RenderProvider       RenderProvider
	StateProvider        StateProvider
	SSOStateCodec        SSOStateCodec
	Interactions         InteractionFlow
	OAuthProviderFactory OAuthProviderFactory
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

func (p *AuthenticateProviderImpl) get(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType) (writeResponse func(err error), err error) {
	var state *State
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		if anyError == nil && state != nil {
			anyError = state.Error
		}
		p.RenderProvider.WritePage(w, r, templateType, anyError)
	}

	state, err = p.StateProvider.RestoreState(r, true)
	if errors.Is(err, ErrStateNotFound) {
		err = nil
	}
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *AuthenticateProviderImpl) handleResult(w http.ResponseWriter, r *http.Request, result *interactionflows.WebAppResult, err error) {
	if err != nil {
		RedirectToCurrentPath(w, r)
		return
	}

	for _, cookie := range result.Cookies {
		httputil.UpdateCookie(w, cookie)
	}

	switch result.Step {
	case interactionflows.WebAppStepAuthenticatePassword:
		RedirectToPathWithX(w, r, "/enter_password")
	case interactionflows.WebAppStepSetupPassword:
		RedirectToPathWithX(w, r, "/create_password")
	case interactionflows.WebAppStepAuthenticateOOBOTP:
		RedirectToPathWithX(w, r, "/oob_otp")
	case interactionflows.WebAppStepSetupOOBOTP:
		RedirectToPathWithX(w, r, "/oob_otp")
	case interactionflows.WebAppStepCompleted:
		RedirectToRedirectURI(w, r, p.ServerConfig.TrustProxy)
	}
}

func (p *AuthenticateProviderImpl) GetLoginForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUILoginHTML)
}

func (p *AuthenticateProviderImpl) LoginWithLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.StateProvider.CreateState(r, err)
		p.handleResult(w, r, result, err)
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDEnterLoginIDRequest, r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	result, err = p.Interactions.LoginWithLoginID(r.Form.Get("x_login_id"))
	if err != nil {
		return
	}

	r.Form["x_interaction_token"] = []string{result.Token}
	return
}

func (p *AuthenticateProviderImpl) GetEnterPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUIEnterPasswordHTML)
}

func (p *AuthenticateProviderImpl) GetOOBOTPForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUIOOBOTPHTML)
}

func (p *AuthenticateProviderImpl) EnterSecret(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		r.Form.Del("x_password")
		p.StateProvider.UpdateState(r, err)
		p.handleResult(w, r, result, err)
	}

	_, err = p.StateProvider.RestoreState(r, false)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDEnterPasswordRequest, r.Form)
	if err != nil {
		return
	}

	result, err = p.Interactions.EnterSecret(
		r.Form.Get("x_interaction_token"),
		r.Form.Get("x_password"),
	)
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) TriggerOOBOTP(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.handleResult(w, r, result, err)
	}

	p.ValidateProvider.PrepareValues(r.Form)

	result, err = p.Interactions.TriggerOOBOTP(r.Form.Get("x_interaction_token"))
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) GetCreateLoginIDForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUISignupHTML)
}

func (p *AuthenticateProviderImpl) GetCreatePasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUICreatePasswordHTML)
}

func (p *AuthenticateProviderImpl) CreateLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.StateProvider.CreateState(r, err)
		p.handleResult(w, r, result, err)
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDCreateLoginIDRequest, r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	result, err = p.Interactions.SignupWithLoginID(
		r.Form.Get("x_login_id_key"),
		r.Form.Get("x_login_id"),
	)
	if err != nil {
		return
	}

	r.Form["x_interaction_token"] = []string{result.Token}
	return
}

func (p *AuthenticateProviderImpl) GetPromoteLoginIDForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUIPromoteHTML)
}

func (p *AuthenticateProviderImpl) PromoteLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.StateProvider.UpdateState(r, err)
		p.handleResult(w, r, result, err)
	}

	state, err := p.StateProvider.RestoreState(r, false)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDCreateLoginIDRequest, r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	result, err = p.Interactions.PromoteWithLoginID(
		r.Form.Get("x_login_id_key"),
		r.Form.Get("x_login_id"),
		state.AnonymousUserID,
	)
	if err != nil {
		return
	}

	r.Form["x_interaction_token"] = []string{result.Token}
	return
}

func (p *AuthenticateProviderImpl) SetLoginID(r *http.Request) (err error) {
	if r.Form.Get("x_login_id_input_type") == "phone" {
		e164, e := phone.Parse(r.Form.Get("x_national_number"), r.Form.Get("x_calling_code"))
		if e != nil {
			err = &validation.AggregatedError{
				Errors: []validation.Error{{
					Keyword:  "format",
					Location: "/x_national_number",
					Info:     map[string]interface{}{},
				}},
			}
			return
		}
		r.Form.Set("x_login_id", e164)
	}

	return
}

func (p *AuthenticateProviderImpl) LoginIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
	var authURI string
	writeResponse = func(err error) {
		p.StateProvider.UpdateState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			http.Redirect(w, r, authURI, http.StatusFound)
		}
	}

	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	p.StateProvider.CreateState(r, nil)

	// set hashed csrf cookies to sso state
	// callback will verify if the request has the same cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
	}
	hashedNonce := crypto.SHA256String(cookie.Value)
	webappSSOState := SSOState{}
	// Redirect back to the current page on error.
	q := r.URL.Query()
	q.Set("error_uri", r.URL.Path)
	webappSSOState.SetRequestQuery(q.Encode())
	state := sso.State{
		Action:      "login",
		HashedNonce: hashedNonce,
		Extra:       webappSSOState,
	}
	encodedState, err := p.SSOStateCodec.EncodeState(state)
	if err != nil {
		return
	}
	authURI, err = oauthProvider.GetAuthURL(state, encodedState)
	return
}

func (p *AuthenticateProviderImpl) GetSettingsIdentity(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUISettingsIdentityHTML)
}

func (p *AuthenticateProviderImpl) LinkIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
	var authURI string
	writeResponse = func(err error) {
		p.StateProvider.UpdateState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			http.Redirect(w, r, authURI, http.StatusFound)
		}
	}

	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

	p.StateProvider.CreateState(r, nil)

	// set hashed csrf cookies to sso state
	// callback will verify if the request has the same cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
	}
	hashedNonce := crypto.SHA256String(cookie.Value)
	webappSSOState := SSOState{}
	// Redirect back to the current page.
	q := r.URL.Query()
	q.Set("redirect_uri", r.URL.Path)
	q.Set("error_uri", r.URL.Path)
	webappSSOState.SetRequestQuery(q.Encode())
	state := sso.State{
		Action:      "link",
		UserID:      userID,
		HashedNonce: hashedNonce,
		Extra:       webappSSOState,
	}
	encodedState, err := p.SSOStateCodec.EncodeState(state)
	if err != nil {
		return
	}
	authURI, err = oauthProvider.GetAuthURL(state, encodedState)
	return
}

func (p *AuthenticateProviderImpl) PromoteIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
	var authURI string
	writeResponse = func(err error) {
		p.StateProvider.UpdateState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			http.Redirect(w, r, authURI, http.StatusFound)
		}
	}

	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	webappState, err := p.StateProvider.RestoreState(r, false)
	if err != nil {
		return
	}

	// set hashed csrf cookies to sso state
	// callback will verify if the request has the same cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
	}
	hashedNonce := crypto.SHA256String(cookie.Value)
	webappSSOState := SSOState{}
	// Redirect back to the current page on error.
	q := r.URL.Query()
	q.Set("error_uri", r.URL.Path)
	webappSSOState.SetRequestQuery(q.Encode())
	state := sso.State{
		Action:      "promote",
		UserID:      webappState.AnonymousUserID,
		HashedNonce: hashedNonce,
		Extra:       webappSSOState,
	}
	encodedState, err := p.SSOStateCodec.EncodeState(state)
	if err != nil {
		return
	}
	authURI, err = oauthProvider.GetAuthURL(state, encodedState)
	return
}

func (p *AuthenticateProviderImpl) UnlinkIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.StateProvider.CreateState(r, err)
		p.handleResult(w, r, result, err)
	}

	providerConfig, ok := p.SSOOAuthConfig.GetProviderConfig(providerAlias)
	if !ok {
		err = ErrOAuthProviderNotFound
		return
	}

	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

	r.Form.Set("redirect_uri", r.URL.Path)

	result, err = p.Interactions.UnlinkWithOAuthProvider(userID, providerConfig)
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) AddOrChangeLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
	writeResponse = func(err error) {
		p.StateProvider.CreateState(r, err)
		RedirectToPathWithX(w, r, "/enter_login_id")
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDAddOrChangeLoginIDRequest, r.Form)
	if err != nil {
		return
	}

	r.Form.Set("redirect_uri", r.URL.Path)

	return
}

func (p *AuthenticateProviderImpl) GetEnterLoginIDForm(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUIEnterLoginIDHTML)
}

func (p *AuthenticateProviderImpl) EnterLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.StateProvider.UpdateState(r, err)
		p.handleResult(w, r, result, err)
	}

	_, err = p.StateProvider.RestoreState(r, false)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDEnterLoginIDRequest, r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

	oldLoginID := r.Form.Get("x_old_login_id_value")
	if oldLoginID != "" {
		result, err = p.Interactions.UpdateLoginID(
			userID,
			loginid.LoginID{
				Key:   r.Form.Get("x_login_id_key"),
				Value: oldLoginID,
			},
			loginid.LoginID{
				Key:   r.Form.Get("x_login_id_key"),
				Value: r.Form.Get("x_login_id"),
			},
		)
	} else {
		result, err = p.Interactions.AddLoginID(userID, loginid.LoginID{
			Key:   r.Form.Get("x_login_id_key"),
			Value: r.Form.Get("x_login_id"),
		})
	}
	if err != nil {
		return
	}

	r.Form["x_interaction_token"] = []string{result.Token}
	return
}

func (p *AuthenticateProviderImpl) RemoveLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.StateProvider.UpdateState(r, err)
		p.handleResult(w, r, result, err)
	}

	_, err = p.StateProvider.RestoreState(r, false)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDRemoveLoginIDRequest, r.Form)
	if err != nil {
		return
	}

	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

	result, err = p.Interactions.RemoveLoginID(userID, loginid.LoginID{
		Key:   r.Form.Get("x_login_id_key"),
		Value: r.Form.Get("x_old_login_id_value"),
	})

	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) HandleSSOCallback(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(error), err error) {
	v := url.Values{}
	writeResponse = func(err error) {
		sid := v.Get("x_sid")

		if err != nil {
			// It is assumed that LoginIdentityProvider and LinkIdentityProvider always
			// generate a state.
			p.StateProvider.UpdateError(sid, err)
			// FIXME: temporary fix, see SkygearIO/skygear-server#1478
			callbackURL := v.Get("error_uri")
			if callbackURL == "" {
				callbackURL = "/login"
			}
			RedirectToPathWithQuery(w, r, callbackURL, v)
		} else {
			callbackURL := v.Get("redirect_uri")
			if callbackURL == "" {
				callbackURL = "/login"
			}
			redirectURI, err := parseRedirectURI(r, callbackURL, false, p.ServerConfig.TrustProxy)
			if err != nil {
				redirectURI = DefaultRedirectURI
			}
			http.Redirect(w, r, redirectURI, http.StatusFound)
		}
	}

	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	err = p.ValidateProvider.Validate(WebAppSchemaIDSSOCallbackRequest, r.Form)
	if err != nil {
		return
	}

	encodedState := r.Form.Get("state")
	state, err := p.SSOStateCodec.DecodeState(encodedState)
	if err != nil {
		return
	}
	webappSSOState := SSOState(state.Extra)
	requestQuery := webappSSOState.RequestQuery()
	v, err = url.ParseQuery(requestQuery)
	if err != nil {
		return writeResponse, &validation.AggregatedError{
			Errors: []validation.Error{{
				Keyword:  "general",
				Location: "/state",
				Info:     map[string]interface{}{},
			}},
		}
	}

	oauthError := r.Form.Get("error")
	if oauthError != "" {
		msg := "login failed"
		if desc := r.Form.Get("error_description"); desc != "" {
			msg += ": " + desc
		}
		err = sso.NewSSOFailed(sso.SSOUnauthorized, msg)
		return
	}

	// verify if the request has the same csrf cookies
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
		return
	}
	hashedCookie := crypto.SHA256String(cookie.Value)
	hashedNonce := state.HashedNonce
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(hashedCookie)) != 1 {
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
		return
	}

	code := r.Form.Get("code")
	scope := r.Form.Get("scope")
	oauthAuthInfo, err := oauthProvider.GetAuthInfo(
		sso.OAuthAuthorizationResponse{
			Code:  code,
			State: encodedState,
			Scope: scope,
		},
		*state,
	)
	if err != nil {
		return
	}

	var result *interactionflows.WebAppResult
	switch state.Action {
	case "login":
		result, err = p.Interactions.LoginWithOAuthProvider(oauthAuthInfo)
	case "link":
		result, err = p.Interactions.LinkWithOAuthProvider(state.UserID, oauthAuthInfo)
	case "promote":
		result, err = p.Interactions.PromoteWithOAuthProvider(state.UserID, oauthAuthInfo)
	}

	if err != nil {
		return
	}

	for _, cookie := range result.Cookies {
		httputil.UpdateCookie(w, cookie)
	}

	return
}
