package webapp

import (
	"net/http"
)

type AuthenticateProviderImpl struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
}

var _ AuthenticateProvider = &AuthenticateProviderImpl{}

func (p *AuthenticateProviderImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		// TODO(webapp): render err in html
		panic(err)
	}

	p.ValidateProvider.Prevalidate(r.Form)

	var writeResponse func(err error)
	var err error
	step := r.Form.Get("x_step")
	switch step {
	case "submit_password":
		writeResponse, err = p.SubmitPassword(w, r)
	case "submit_login_id":
		writeResponse, err = p.SubmitLoginID(w, r)
	case "choose_idp":
		writeResponse, err = p.ChooseIdentityProvider(w, r)
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

func (p *AuthenticateProviderImpl) SubmitLoginID(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	err = p.ValidateProvider.Validate("#WebAppAuthenticateLoginIDRequest", r.Form)
	writeResponse = func(err error) {
		t := TemplateItemTypeAuthUISignInHTML
		if err == nil {
			t = TemplateItemTypeAuthUISignInPasswordHTML
		}
		p.RenderProvider.WritePage(w, r, t, err)
	}
	return
}

func (p *AuthenticateProviderImpl) SubmitPassword(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	// TODO(webapp): Enter the authentication process
	return p.Default(w, r)
}

func (p *AuthenticateProviderImpl) ChooseIdentityProvider(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	// TODO(webapp): Prepare IdP authorization URL and respond 302
	// TODO(webapp): Add a new endpoint to be redirect_uri
	return p.Default(w, r)
}
