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
	var writeResponse func(err error)
	var err error
	writeResponse, err = p.Default(w, r)
	writeResponse(err)
}

func (p *AuthenticateProviderImpl) Default(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	p.ValidateProvider.Prevalidate(r.Form)
	err = p.ValidateProvider.Validate("#WebAppAuthenticateRequest", r.Form)
	writeResponse = func(err error) {
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUISignInHTML, err)
	}
	return
}
