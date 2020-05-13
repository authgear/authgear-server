package webapp

import (
	"crypto/subtle"
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type InteractionFlow interface {
	LoginWithLoginID(loginID string) (*interactionflows.WebAppResult, error)
	SignupWithLoginID(loginIDKey, loginID string) (*interactionflows.WebAppResult, error)
	AuthenticateSecret(token string, secret string) (*interactionflows.WebAppResult, error)
	TriggerOOBOTP(token string, step interaction.Step) (*interactionflows.WebAppResult, error)
	SetupSecret(token string, secret string) (*interactionflows.WebAppResult, error)
	LoginWithOAuthProvider(oauthAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
}

type AuthenticateProviderImpl struct {
	ValidateProvider     ValidateProvider
	RenderProvider       RenderProvider
	AuthnProvider        AuthnProvider
	StateStore           StateStore
	SSOProvider          sso.Provider
	Interactions         InteractionFlow
	OAuthProviderFactory OAuthProviderFactory
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
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

func (p *AuthenticateProviderImpl) makeState(sid string) *State {
	s, err := p.StateStore.Get(sid)
	if err != nil {
		s = NewState()
	}

	return s
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

	state, err = p.restoreState(r)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *AuthenticateProviderImpl) GetEnterLoginIDForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	return p.get(w, r, TemplateItemTypeAuthUILoginHTML)
}

func (p *AuthenticateProviderImpl) EnterLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var result *interactionflows.WebAppResult
	writeResponse = func(err error) {
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			var nextPath string
			switch result.Step {
			case interactionflows.WebAppStepAuthenticatePassword:
				nextPath = "/enter_password"
			case interactionflows.WebAppStepAuthenticateOOBOTP:
				nextPath = "/oob_otp"
			default:
				panic("interaction_flow_webapp: unexpected step " + result.Step)
			}
			RedirectToPathWithX(w, r, nextPath)
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppEnterLoginIDRequest", r.Form)
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

	err = p.ValidateProvider.Validate("#WebAppEnterPasswordRequest", r.Form)
	if err != nil {
		return
	}

	// TODO(interaction): make all handler to call .handleResult
	// to write interaction token to r.Form and
	// and set cookies to w.
	// It is not harmful to always do that.
	result, err := p.Interactions.AuthenticateSecret(
		r.Form.Get("x_interaction_token"),
		r.Form.Get("x_password"),
	)
	if err != nil {
		return
	}

	for _, cookie := range result.Cookies {
		corehttp.UpdateCookie(w, cookie)
	}

	return
}

func (p *AuthenticateProviderImpl) TriggerOOBOTP(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		r.Form.Del("x_password")
		p.persistState(r, err)
		RedirectToCurrentPath(w, r)
	}

	p.ValidateProvider.PrepareValues(r.Form)

	result, err := p.Interactions.TriggerOOBOTP(r.Form.Get("x_interaction_token"), interaction.StepAuthenticatePrimary)
	if err != nil {
		return
	}

	r.Form["x_interaction_token"] = []string{result.Token}

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
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			var nextPath string
			switch result.Step {
			case interactionflows.WebAppStepSetupPassword:
				nextPath = "/create_password"
			case interactionflows.WebAppStepSetupOOBOTP:
				nextPath = "/oob_otp"
			default:
				panic("interaction_flow_webapp: unexpected step " + result.Step)
			}
			RedirectToPathWithX(w, r, nextPath)
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppCreateLoginIDRequest", r.Form)
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

func (p *AuthenticateProviderImpl) CreateSecret(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
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

	err = p.ValidateProvider.Validate("#WebAppEnterPasswordRequest", r.Form)
	if err != nil {
		return
	}

	result, err := p.Interactions.SetupSecret(
		r.Form.Get("x_interaction_token"),
		r.Form.Get("x_password"),
	)
	if err != nil {
		return
	}

	for _, cookie := range result.Cookies {
		corehttp.UpdateCookie(w, cookie)
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

func (p *AuthenticateProviderImpl) LoginIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
	var authURI string
	writeResponse = func(err error) {
		p.persistState(r, err)
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

	// create or update ui state
	// state id will be set into the request query
	p.persistState(r, nil)

	// set hashed csrf cookies to sso state
	// callback will verify if the request has the same cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
	}
	hashedNonce := crypto.SHA256String(cookie.Value)
	webappSSOState := SSOState{}
	webappSSOState.SetRequestQuery(r.URL.Query().Encode())
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

func (p *AuthenticateProviderImpl) HandleSSOCallback(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(error), err error) {
	v := url.Values{}
	writeResponse = func(err error) {
		callbackURL := v.Get("redirect_uri")
		if callbackURL == "" {
			callbackURL = "/"
		}
		sid := v.Get("x_sid")

		if err != nil {
			// try to obtain state id from sso state
			// create new state if failed
			s := p.makeState(sid)
			s.SetError(err)
			if e := p.StateStore.Set(s); e != nil {
				panic(e)
			}
			// x_sid maybe new if callback failed to obtain the state
			v.Set("x_sid", s.ID)
			RedirectToPathWithQuery(w, r, "/login", v)
		} else {
			redirectURI, err := parseRedirectURI(r, callbackURL)
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
	requestQuery := webappSSOState.RequestQuery()
	v, err = url.ParseQuery(requestQuery)
	if err != nil {
		return writeResponse, validation.NewValidationFailed("", []validation.ErrorCause{
			validation.ErrorCause{
				Kind:    validation.ErrorGeneral,
				Pointer: "/state",
			},
		})
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
	if state.Action == "login" {
		result, err = p.Interactions.LoginWithOAuthProvider(oauthAuthInfo)
	} else {
		panic("only login is supported")
	}

	if err != nil {
		return
	}

	for _, cookie := range result.Cookies {
		corehttp.UpdateCookie(w, cookie)
	}

	return
}
