package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type AuthenticateProviderImpl struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
	AuthnProvider    AuthnProvider
	StateStore       StateStore
	SSOProvider      sso.Provider
}

type AuthnProvider interface {
	LoginWithLoginID(
		client config.OAuthClientConfiguration,
		loginID loginid.LoginID,
		plainPassword string,
	) (authn.Result, error)

	ValidateSignupLoginID(loginid loginid.LoginID) error

	SignupWithLoginIDs(
		client config.OAuthClientConfiguration,
		loginIDs []loginid.LoginID,
		plainPassword string,
		metadata map[string]interface{},
		onUserDuplicate model.OnUserDuplicate,
	) (authn.Result, error)

	WriteCookie(rw http.ResponseWriter, result *authn.CompletionResult)

	OAuthAuthenticate(
		client config.OAuthClientConfiguration,
		authInfo sso.AuthInfo,
		loginState sso.LoginState,
	) (authn.Result, error)
}

type OAuthProvider interface {
	GetAuthURL(state sso.State, encodedState string) (url string, err error)
	GetAuthInfo(r sso.OAuthAuthorizationResponse, state sso.State) (sso.AuthInfo, error)
}

func (p *AuthenticateProviderImpl) persistState(r *http.Request, inputError error) {
	s, err := p.StateStore.Get(r.URL.Query().Get("x_sid"))
	if err != nil {
		s = NewState()
		q := r.URL.Query()
		q.Set("x_sid", s.ID)
		r.URL.RawQuery = q.Encode()
	}

	s.SetForm(r.Form)
	s.SetError(inputError)

	err = p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}
}

func (p *AuthenticateProviderImpl) restoreState(r *http.Request) (state *State, err error) {
	state, err = p.StateStore.Get(r.URL.Query().Get("x_sid"))
	if err != nil {
		if err == ErrStateNotFound {
			err = nil
		}
		return
	}
	err = state.Restore(r.Form)
	if err != nil {
		return
	}
	return state, nil
}

func (p *AuthenticateProviderImpl) get(w http.ResponseWriter, r *http.Request, schemaName string, templateType config.TemplateItemType) (writeResponse func(err error), err error) {
	var state *State
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		if anyError == nil && state != nil {
			anyError = state.Error
		}
		p.RenderProvider.WritePage(w, r, templateType, anyError)
	}

	state, err = p.restoreState(r)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(schemaName, r.Form)
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) GetLoginForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, "#WebAppLoginRequest", TemplateItemTypeAuthUILoginHTML)
}

func (p *AuthenticateProviderImpl) PostLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToPathWithQueryPreserved(w, r, "/login/password")
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppLoginLoginIDRequest", r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) GetLoginPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, "#WebAppLoginLoginIDRequest", TemplateItemTypeAuthUILoginPasswordHTML)
}

func (p *AuthenticateProviderImpl) PostLoginPassword(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		r.Form.Del("x_password")
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToRedirectURI(w, r)
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppLoginLoginIDPasswordRequest", r.Form)
	if err != nil {
		return
	}

	var client config.OAuthClientConfiguration
	loginID := loginid.LoginID{Value: r.Form.Get("x_login_id")}
	result, err := p.AuthnProvider.LoginWithLoginID(client, loginID, r.Form.Get("x_password"))
	if err != nil {
		return
	}

	switch r := result.(type) {
	case *authn.CompletionResult:
		p.AuthnProvider.WriteCookie(w, r)
	case *authn.InProgressResult:
		panic("TODO(webapp): handle MFA")
	}

	return
}

func (p *AuthenticateProviderImpl) GetSignupForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, "#WebAppSignupRequest", TemplateItemTypeAuthUISignupHTML)
}

func (p *AuthenticateProviderImpl) GetSignupPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, "#WebAppSignupLoginIDRequest", TemplateItemTypeAuthUISignupPasswordHTML)
}

func (p *AuthenticateProviderImpl) PostSignupLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToPathWithQueryPreserved(w, r, "/signup/password")
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppSignupLoginIDRequest", r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	err = p.AuthnProvider.ValidateSignupLoginID(loginid.LoginID{
		Key:   r.Form.Get("x_login_id_key"),
		Value: r.Form.Get("x_login_id"),
	})
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) PostSignupPassword(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		r.Form.Del("x_password")
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToRedirectURI(w, r)
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppSignupLoginIDPasswordRequest", r.Form)
	if err != nil {
		return
	}

	var client config.OAuthClientConfiguration
	result, err := p.AuthnProvider.SignupWithLoginIDs(
		client,
		[]loginid.LoginID{
			loginid.LoginID{
				Key:   r.Form.Get("x_login_id_key"),
				Value: r.Form.Get("x_login_id"),
			},
		},
		r.Form.Get("x_password"), map[string]interface{}{},
		model.OnUserDuplicateAbort,
	)
	if err != nil {
		return
	}

	switch r := result.(type) {
	case *authn.CompletionResult:
		p.AuthnProvider.WriteCookie(w, r)
	case *authn.InProgressResult:
		panic("TODO(webapp): handle MFA")
	}

	return
}

func (p *AuthenticateProviderImpl) SetLoginID(r *http.Request) (err error) {
	if r.Form.Get("x_login_id_input_type") == "phone" {
		e164, e := phone.Parse(r.Form.Get("x_national_number"), r.Form.Get("x_calling_code"))
		if e != nil {
			err = validation.NewValidationFailed("", []validation.ErrorCause{
				validation.ErrorCause{
					Kind:    validation.ErrorStringFormat,
					Pointer: "/x_national_number",
				},
			})
			return
		}
		r.Form.Set("x_login_id", e164)
	}

	return
}

func (p *AuthenticateProviderImpl) ChooseIdentityProvider(w http.ResponseWriter, r *http.Request, oauthProvider OAuthProvider) (writeResponse func(err error), err error) {
	var authURI string
	writeResponse = func(err error) {
		if err != nil {
			panic(err)
		} else {
			http.Redirect(w, r, authURI, http.StatusFound)
		}
	}

	redirectURI := "/"
	if u, err := getRedirectURI(r); err == nil {
		redirectURI = u
	}
	nonce := sso.GenerateOpenIDConnectNonce()
	hashedNonce := crypto.SHA256String(nonce)
	webappSSOState := SSOState{}
	webappSSOState.SetCallbackURL(redirectURI)
	state := sso.State{
		Action: "login",
		LoginState: sso.LoginState{
			OnUserDuplicate: model.OnUserDuplicateAbort,
		},
		HashedNonce: hashedNonce,
		Extra:       webappSSOState,
	}
	encodedState, err := p.SSOProvider.EncodeState(state)
	if err != nil {
		return
	}
	authURI, err = oauthProvider.GetAuthURL(state, encodedState)
	return
}

func (p *AuthenticateProviderImpl) HandleSSOCallback(w http.ResponseWriter, r *http.Request, oauthProvider OAuthProvider) (writeResponse func(error), err error) {
	var callbackURL string
	writeResponse = func(err error) {
		if err != nil {
			panic(err)
		} else {
			redirectURI, err := parseRedirectURI(r, callbackURL)
			if err != nil {
				redirectURI = DefaultRedirectURI
			}
			http.Redirect(w, r, redirectURI, http.StatusFound)
		}
	}

	err = p.ValidateProvider.Validate("#SSOCallbackRequest", r.Form)
	if err != nil {
		return
	}

	code := r.Form.Get("code")
	encodedState := r.Form.Get("state")
	scope := r.Form.Get("scope")
	state, err := p.SSOProvider.DecodeState(encodedState)
	if err != nil {
		return
	}
	webappSSOState := SSOState(state.Extra)
	callbackURL = webappSSOState.CallbackURL()
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

	var result authn.Result
	if state.Action == "login" {
		// TODO(webapp): link provider
		var client config.OAuthClientConfiguration
		result, err = p.AuthnProvider.OAuthAuthenticate(
			client,
			oauthAuthInfo,
			state.LoginState,
		)
	} else {
		panic("only login is supported")
	}

	if err != nil {
		return
	}

	switch r := result.(type) {
	case *authn.CompletionResult:
		p.AuthnProvider.WriteCookie(w, r)
	case *authn.InProgressResult:
		panic("TODO(webapp): handle MFA")
	}

	return
}
