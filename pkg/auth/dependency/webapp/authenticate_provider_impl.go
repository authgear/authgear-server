package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type AuthenticateProviderImpl struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
	AuthnProvider    AuthnProvider
}

var _ AuthenticateProvider = &AuthenticateProviderImpl{}

type AuthnProvider interface {
	LoginWithLoginID(
		client config.OAuthClientConfiguration,
		loginID loginid.LoginID,
		plainPassword string,
	) (authn.Result, error)

	ValidateSignUpLoginID(loginid loginid.LoginID) error

	WriteCookie(rw http.ResponseWriter, result *authn.CompletionResult)
}

func (p *AuthenticateProviderImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	var writeResponse func(err error)
	var err error
	step := r.Form.Get("x_step")
	switch step {
	case "submit_login_id":
		writeResponse, err = p.SubmitLoginID(w, r)
	case "submit_password":
		writeResponse, err = p.SubmitPassword(w, r)
	case "choose_idp":
		writeResponse, err = p.ChooseIdentityProvider(w, r)
	case "sign_up":
		writeResponse, err = p.SignUp(w, r)
	case "sign_up_submit_login_id":
		writeResponse, err = p.SignUpSubmitLoginID(w, r)
	default:
		writeResponse, err = p.Default(w, r)
	}
	writeResponse(err)
}

func (p *AuthenticateProviderImpl) Default(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	err = p.ValidateProvider.Validate("#WebAppAuthenticateRequest", r.Form)
	writeResponse = func(err error) {
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUISignInHTML, err)
	}
	return
}

func (p *AuthenticateProviderImpl) SignUp(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	err = p.ValidateProvider.Validate("#WebAppSignUpRequest", r.Form)
	writeResponse = func(err error) {
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUISignUpHTML, err)
	}
	return
}

func (p *AuthenticateProviderImpl) SignUpSubmitLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		t := TemplateItemTypeAuthUISignUpHTML
		if err == nil {
			t = TemplateItemTypeAuthUISignUpPasswordHTML
		}
		p.RenderProvider.WritePage(w, r, t, err)
	}

	err = p.ValidateProvider.Validate("#WebAppSignUpLoginIDRequest", r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	err = p.AuthnProvider.ValidateSignUpLoginID(loginid.LoginID{
		Key:   r.Form.Get("x_login_id_key"),
		Value: r.Form.Get("x_login_id"),
	})
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) SubmitLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		t := TemplateItemTypeAuthUISignInHTML
		if err == nil {
			t = TemplateItemTypeAuthUISignInPasswordHTML
		}
		p.RenderProvider.WritePage(w, r, t, err)
	}

	err = p.ValidateProvider.Validate("#WebAppAuthenticateLoginIDRequest", r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
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

func (p *AuthenticateProviderImpl) SubmitPassword(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		if err != nil {
			t := TemplateItemTypeAuthUISignInPasswordHTML
			p.RenderProvider.WritePage(w, r, t, err)
		} else {
			RedirectToRedirectURI(w, r)
		}
	}

	err = p.ValidateProvider.Validate("#WebAppAuthenticateLoginIDPasswordRequest", r.Form)
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

func (p *AuthenticateProviderImpl) ChooseIdentityProvider(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	// TODO(webapp): Prepare IdP authorization URL and respond 302
	// TODO(webapp): Add a new endpoint to be redirect_uri
	return p.Default(w, r)
}
