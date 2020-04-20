package webapp

import (
	"net/http"
)

type ForgotPasswordProvider struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
}

func (p *ForgotPasswordProvider) GetForgotPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUIForgotPasswordHTML, anyError)
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}
