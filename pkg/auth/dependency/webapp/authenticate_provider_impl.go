package webapp

import (
// "net/url"
// "github.com/authgear/authgear-server/pkg/auth/dependency/auth"
// "github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
// "github.com/authgear/authgear-server/pkg/core/phone"
// "github.com/authgear/authgear-server/pkg/validation"
)

//
// type InteractionFlow interface {
// 	LoginWithLoginID(loginID string) (*interactionflows.WebAppResult, error)
// 	SignupWithLoginID(loginIDKey, loginID string) (*interactionflows.WebAppResult, error)
// 	PromoteWithLoginID(loginIDKey, loginID string, userID string) (*interactionflows.WebAppResult, error)
// 	EnterSecret(i *interaction.Interaction, secret string) (*interactionflows.WebAppResult, error)
// 	TriggerOOBOTP(i *interaction.Interaction) (*interactionflows.WebAppResult, error)
// 	LoginWithOAuthProvider(oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
// 	LinkWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
// 	UnlinkWithOAuthProvider(userID string, providerConfig *config.OAuthSSOProviderConfig) (*interactionflows.WebAppResult, error)
// 	PromoteWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
// 	AddLoginID(userID string, loginID loginid.LoginID) (*interactionflows.WebAppResult, error)
// 	UpdateLoginID(userID string, oldLoginID loginid.LoginID, newLoginID loginid.LoginID) (*interactionflows.WebAppResult, error)
// 	RemoveLoginID(userID string, loginID loginid.LoginID) (*interactionflows.WebAppResult, error)
//
// 	GetInteractionState(i *interaction.Interaction) (*interaction.State, error)
// }
//

//
// type AuthenticateProviderImpl struct {
// 	ServerConfig         *config.ServerConfig
// 	SSOOAuthConfig       *config.OAuthSSOConfig
// 	StateProvider        StateProvider
// 	SSOStateCodec        SSOStateCodec
// 	Interactions         InteractionFlow
// 	OAuthProviderFactory OAuthProviderFactory
// }
//

//
// // func (p *AuthenticateProviderImpl) get(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType) (writeResponse func(err error), err error) {
// // 	var state *State
// // 	writeResponse = func(err error) {
// // 		var anyError interface{}
// // 		anyError = err
// // 		if anyError == nil && state != nil {
// // 			anyError = state.Error
// // 		}
// // 		p.RenderProvider.WritePage(w, r, templateType, anyError)
// // 	}
// //
// // 	state, err = p.StateProvider.RestoreState(r, true)
// // 	if errors.Is(err, ErrStateNotFound) {
// // 		err = nil
// // 	}
// // 	if err != nil {
// // 		return
// // 	}
// //
// // 	p.ValidateProvider.PrepareValues(r.Form)
// //
// // 	return
// // }
//
// func (p *AuthenticateProviderImpl) handleResult(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	state *State,
// 	result *interactionflows.WebAppResult,
// 	err error,
// ) {
// }
//
// func (p *AuthenticateProviderImpl) LoginWithLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
// 	var result *interactionflows.WebAppResult
// 	writeResponse = func(err error) {
// 		s := p.StateProvider.CreateState(r, result, err)
// 		p.handleResult(w, r, s, result, err)
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDEnterLoginIDRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	err = p.SetLoginID(r)
// 	if err != nil {
// 		return
// 	}
//
// 	result, err = p.Interactions.LoginWithLoginID(r.Form.Get("x_login_id"))
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) EnterSecret(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
// 	var state *State
// 	var result *interactionflows.WebAppResult
//
// 	writeResponse = func(err error) {
// 		p.StateProvider.UpdateState(state, result, err)
// 		p.handleResult(w, r, state, result, err)
// 	}
//
// 	state, err = p.StateProvider.RestoreState(r, false)
// 	if err != nil {
// 		return
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDEnterPasswordRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	result, err = p.Interactions.EnterSecret(
// 		state.Interaction,
// 		r.Form.Get("x_password"),
// 	)
//
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) TriggerOOBOTP(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
// 	var result *interactionflows.WebAppResult
// 	var state *State
// 	writeResponse = func(err error) {
// 		p.handleResult(w, r, state, result, err)
// 	}
//
// 	state, err = p.StateProvider.RestoreState(r, false)
// 	if err != nil {
// 		return
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	result, err = p.Interactions.TriggerOOBOTP(state.Interaction)
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) CreateLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
// 	var result *interactionflows.WebAppResult
// 	writeResponse = func(err error) {
// 		s := p.StateProvider.CreateState(r, result, err)
// 		p.handleResult(w, r, s, result, err)
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDCreateLoginIDRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	err = p.SetLoginID(r)
// 	if err != nil {
// 		return
// 	}
//
// 	result, err = p.Interactions.SignupWithLoginID(
// 		r.Form.Get("x_login_id_key"),
// 		r.Form.Get("x_login_id"),
// 	)
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) PromoteLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
// 	var result *interactionflows.WebAppResult
// 	var state *State
//
// 	writeResponse = func(err error) {
// 		p.StateProvider.UpdateState(state, result, err)
// 		p.handleResult(w, r, state, result, err)
// 	}
//
// 	state, err = p.StateProvider.RestoreState(r, false)
// 	if err != nil {
// 		return
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDCreateLoginIDRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	err = p.SetLoginID(r)
// 	if err != nil {
// 		return
// 	}
//
// 	result, err = p.Interactions.PromoteWithLoginID(
// 		r.Form.Get("x_login_id_key"),
// 		r.Form.Get("x_login_id"),
// 		state.AnonymousUserID,
// 	)
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) SetLoginID(r *http.Request) (err error) {
// 	if r.Form.Get("x_login_id_input_type") == "phone" {
// 		e164, e := phone.Parse(r.Form.Get("x_national_number"), r.Form.Get("x_calling_code"))
// 		if e != nil {
// 			err = &validation.AggregatedError{
// 				Errors: []validation.Error{{
// 					Keyword:  "format",
// 					Location: "/x_national_number",
// 					Info:     map[string]interface{}{},
// 				}},
// 			}
// 			return
// 		}
// 		r.Form.Set("x_login_id", e164)
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) LinkIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
// 	var authURI string
// 	var state *State
// 	writeResponse = func(err error) {
// 		p.StateProvider.UpdateState(state, nil, err)
// 		if err != nil {
// 			RedirectToCurrentPath(w, r)
// 		} else {
// 			http.Redirect(w, r, authURI, http.StatusFound)
// 		}
// 	}
//
// 	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
// 	if oauthProvider == nil {
// 		err = ErrOAuthProviderNotFound
// 		return
// 	}
//
// 	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
//
// 	state = p.StateProvider.CreateState(r, nil, nil)
//
// 	// set hashed csrf cookies to sso state
// 	// callback will verify if the request has the same cookie
// 	cookie, err := r.Cookie(csrfCookieName)
// 	if err != nil || cookie.Value == "" {
// 		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
// 	}
// 	hashedNonce := crypto.SHA256String(cookie.Value)
// 	webappSSOState := SSOState{}
// 	// Redirect back to the current page.
// 	q := r.URL.Query()
// 	q.Set("redirect_uri", r.URL.Path)
// 	q.Set("error_uri", r.URL.Path)
// 	webappSSOState.SetRequestQuery(q.Encode())
// 	ssoState := sso.State{
// 		Action:      "link",
// 		UserID:      userID,
// 		HashedNonce: hashedNonce,
// 		Extra:       webappSSOState,
// 	}
// 	encodedState, err := p.SSOStateCodec.EncodeState(ssoState)
// 	if err != nil {
// 		return
// 	}
// 	authURI, err = oauthProvider.GetAuthURL(ssoState, encodedState)
// 	return
// }
//
// func (p *AuthenticateProviderImpl) PromoteIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
// 	var authURI string
// 	var state *State
// 	writeResponse = func(err error) {
// 		p.StateProvider.UpdateState(state, nil, err)
// 		if err != nil {
// 			RedirectToCurrentPath(w, r)
// 		} else {
// 			http.Redirect(w, r, authURI, http.StatusFound)
// 		}
// 	}
//
// 	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
// 	if oauthProvider == nil {
// 		err = ErrOAuthProviderNotFound
// 		return
// 	}
//
// 	state, err = p.StateProvider.RestoreState(r, false)
// 	if err != nil {
// 		return
// 	}
//
// 	// set hashed csrf cookies to sso state
// 	// callback will verify if the request has the same cookie
// 	cookie, err := r.Cookie(csrfCookieName)
// 	if err != nil || cookie.Value == "" {
// 		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
// 	}
// 	hashedNonce := crypto.SHA256String(cookie.Value)
// 	webappSSOState := SSOState{}
// 	// Redirect back to the current page on error.
// 	q := r.URL.Query()
// 	q.Set("error_uri", r.URL.Path)
// 	webappSSOState.SetRequestQuery(q.Encode())
// 	ssoState := sso.State{
// 		Action:      "promote",
// 		UserID:      state.AnonymousUserID,
// 		HashedNonce: hashedNonce,
// 		Extra:       webappSSOState,
// 	}
// 	encodedState, err := p.SSOStateCodec.EncodeState(ssoState)
// 	if err != nil {
// 		return
// 	}
// 	authURI, err = oauthProvider.GetAuthURL(ssoState, encodedState)
// 	return
// }
//
// func (p *AuthenticateProviderImpl) UnlinkIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
// 	var result *interactionflows.WebAppResult
// 	writeResponse = func(err error) {
// 		s := p.StateProvider.CreateState(r, result, err)
// 		p.handleResult(w, r, s, result, err)
// 	}
//
// 	providerConfig, ok := p.SSOOAuthConfig.GetProviderConfig(providerAlias)
// 	if !ok {
// 		err = ErrOAuthProviderNotFound
// 		return
// 	}
//
// 	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
//
// 	r.Form.Set("redirect_uri", r.URL.Path)
//
// 	result, err = p.Interactions.UnlinkWithOAuthProvider(userID, providerConfig)
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) AddOrChangeLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
// 	writeResponse = func(err error) {
// 		p.StateProvider.CreateState(r, nil, err)
// 		RedirectToPathWithX(w, r, "/enter_login_id")
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDAddOrChangeLoginIDRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	r.Form.Set("redirect_uri", r.URL.Path)
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) EnterLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
// 	var result *interactionflows.WebAppResult
// 	var state *State
// 	writeResponse = func(err error) {
// 		p.StateProvider.UpdateState(state, result, err)
// 		p.handleResult(w, r, state, result, err)
// 	}
//
// 	state, err = p.StateProvider.RestoreState(r, false)
// 	if err != nil {
// 		return
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDEnterLoginIDRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	err = p.SetLoginID(r)
// 	if err != nil {
// 		return
// 	}
//
// 	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
//
// 	oldLoginID := r.Form.Get("x_old_login_id_value")
// 	if oldLoginID != "" {
// 		result, err = p.Interactions.UpdateLoginID(
// 			userID,
// 			loginid.LoginID{
// 				Key:   r.Form.Get("x_login_id_key"),
// 				Value: oldLoginID,
// 			},
// 			loginid.LoginID{
// 				Key:   r.Form.Get("x_login_id_key"),
// 				Value: r.Form.Get("x_login_id"),
// 			},
// 		)
// 	} else {
// 		result, err = p.Interactions.AddLoginID(userID, loginid.LoginID{
// 			Key:   r.Form.Get("x_login_id_key"),
// 			Value: r.Form.Get("x_login_id"),
// 		})
// 	}
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) RemoveLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(error), err error) {
// 	var result *interactionflows.WebAppResult
// 	var state *State
// 	writeResponse = func(err error) {
// 		p.StateProvider.UpdateState(state, result, err)
// 		p.handleResult(w, r, state, result, err)
// 	}
//
// 	state, err = p.StateProvider.RestoreState(r, false)
// 	if err != nil {
// 		return
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDRemoveLoginIDRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
//
// 	result, err = p.Interactions.RemoveLoginID(userID, loginid.LoginID{
// 		Key:   r.Form.Get("x_login_id_key"),
// 		Value: r.Form.Get("x_old_login_id_value"),
// 	})
//
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// func (p *AuthenticateProviderImpl) HandleSSOCallback(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(error), err error) {
// 	v := url.Values{}
// 	writeResponse = func(err error) {
// 		// sid := v.Get("x_sid")
//
// 		if err != nil {
// 			// It is assumed that LoginIdentityProvider and LinkIdentityProvider always
// 			// generate a state.
// 			// TODO: Fix SSO state
// 			// FIXME: temporary fix, see SkygearIO/skygear-server#1478
// 			// p.StateProvider.UpdateError(sid, err)
// 			callbackURL := v.Get("error_uri")
// 			if callbackURL == "" {
// 				callbackURL = "/login"
// 			}
// 			RedirectToPathWithQuery(w, r, callbackURL, v)
// 		} else {
// 			callbackURL := v.Get("redirect_uri")
// 			if callbackURL == "" {
// 				callbackURL = "/login"
// 			}
// 			redirectURI, err := parseRedirectURI(r, callbackURL, false, p.ServerConfig.TrustProxy)
// 			if err != nil {
// 				redirectURI = DefaultRedirectURI
// 			}
// 			http.Redirect(w, r, redirectURI, http.StatusFound)
// 		}
// 	}
//
// 	oauthProvider := p.OAuthProviderFactory.NewOAuthProvider(providerAlias)
// 	if oauthProvider == nil {
// 		err = ErrOAuthProviderNotFound
// 		return
// 	}
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDSSOCallbackRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	encodedState := r.Form.Get("state")
// 	state, err := p.SSOStateCodec.DecodeState(encodedState)
// 	if err != nil {
// 		return
// 	}
// 	webappSSOState := SSOState(state.Extra)
// 	requestQuery := webappSSOState.RequestQuery()
// 	v, err = url.ParseQuery(requestQuery)
// 	if err != nil {
// 		return writeResponse, &validation.AggregatedError{
// 			Errors: []validation.Error{{
// 				Keyword:  "general",
// 				Location: "/state",
// 				Info:     map[string]interface{}{},
// 			}},
// 		}
// 	}
//
// 	oauthError := r.Form.Get("error")
// 	if oauthError != "" {
// 		msg := "login failed"
// 		if desc := r.Form.Get("error_description"); desc != "" {
// 			msg += ": " + desc
// 		}
// 		err = sso.NewSSOFailed(sso.SSOUnauthorized, msg)
// 		return
// 	}
//
// 	// verify if the request has the same csrf cookies
// 	cookie, err := r.Cookie(csrfCookieName)
// 	if err != nil || cookie.Value == "" {
// 		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
// 		return
// 	}
// 	hashedCookie := crypto.SHA256String(cookie.Value)
// 	hashedNonce := state.HashedNonce
// 	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(hashedCookie)) != 1 {
// 		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
// 		return
// 	}
//
// 	code := r.Form.Get("code")
// 	scope := r.Form.Get("scope")
// 	oauthAuthInfo, err := oauthProvider.GetAuthInfo(
// 		sso.OAuthAuthorizationResponse{
// 			Code:  code,
// 			State: encodedState,
// 			Scope: scope,
// 		},
// 		*state,
// 	)
// 	if err != nil {
// 		return
// 	}
//
// 	var result *interactionflows.WebAppResult
// 	switch state.Action {
// 	case "login":
// 		result, err = p.Interactions.LoginWithOAuthProvider(oauthAuthInfo)
// 	case "link":
// 		result, err = p.Interactions.LinkWithOAuthProvider(state.UserID, oauthAuthInfo)
// 	case "promote":
// 		result, err = p.Interactions.PromoteWithOAuthProvider(state.UserID, oauthAuthInfo)
// 	}
//
// 	if err != nil {
// 		return
// 	}
//
// 	for _, cookie := range result.Cookies {
// 		httputil.UpdateCookie(w, cookie)
// 	}
//
// 	return
// }
