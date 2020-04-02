package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type AuthenticateProviderImpl struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
	AuthnProvider    AuthnProvider
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

func (p *AuthenticateProviderImpl) GetLoginForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	p.ValidateProvider.PrepareValues(r.Form)
	err = p.ValidateProvider.Validate("#WebAppLoginRequest", r.Form)
	writeResponse = func(err error) {
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUILoginHTML, err)
	}
	return
}

func (p *AuthenticateProviderImpl) PostLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		if err != nil {
			// TODO(webapp): store err in cookie
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToPathWithQueryPreserved(w, r, "/login/password")
		}
	}

	err = p.ValidateProvider.Validate("#WebAppLoginLoginIDRequest", r.Form)
	if err != nil {
		return
	}

	// TODO(webapp): store x_login_id in cookie
	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	return
}

func (p *AuthenticateProviderImpl) GetLoginPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	p.ValidateProvider.PrepareValues(r.Form)
	err = p.ValidateProvider.Validate("#WebAppLoginLoginIDRequest", r.Form)
	writeResponse = func(err error) {
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUILoginPasswordHTML, err)
	}
	return
}

func (p *AuthenticateProviderImpl) PostLoginPassword(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		if err != nil {
			// TODO(webapp): store err in cookie
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToRedirectURI(w, r)
		}
	}

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

func (p *AuthenticateProviderImpl) SignUp(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	err = p.ValidateProvider.Validate("#WebAppSignupRequest", r.Form)
	writeResponse = func(err error) {
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUISignupHTML, err)
	}
	return
}

func (p *AuthenticateProviderImpl) SignUpSubmitLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		t := TemplateItemTypeAuthUISignupHTML
		if err == nil {
			t = TemplateItemTypeAuthUISignupPasswordHTML
		}
		p.RenderProvider.WritePage(w, r, t, err)
	}

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

func (p *AuthenticateProviderImpl) SignUpSubmitPassword(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		if err != nil {
			t := TemplateItemTypeAuthUISignupPasswordHTML
			p.RenderProvider.WritePage(w, r, t, err)
		} else {
			RedirectToRedirectURI(w, r)
		}
	}

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
