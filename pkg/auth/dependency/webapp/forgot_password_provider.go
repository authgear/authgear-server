package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/core/phone"
	"github.com/authgear/authgear-server/pkg/validation"
)

type ForgotPassword interface {
	SendCode(loginID string) error
	ResetPassword(code string, newPassword string) error
}

type ForgotPasswordProvider struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
	StateProvider    StateProvider
	ForgotPassword   ForgotPassword
}

func (p *ForgotPasswordProvider) GetForgotPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var state *State
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		if anyError == nil && state != nil {
			anyError = state.Error
		}
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUIForgotPasswordHTML, anyError)
	}

	state, err = p.StateProvider.RestoreState(r, true)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *ForgotPasswordProvider) PostForgotPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		p.StateProvider.CreateState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToPathWithX(w, r, "/forgot_password/success")
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDForgotPasswordRequest, r.Form)
	if err != nil {
		return
	}

	err = p.SetLoginID(r)
	if err != nil {
		return
	}

	err = p.ForgotPassword.SendCode(r.Form.Get("x_login_id"))
	if err != nil {
		return
	}

	return
}

func (p *ForgotPasswordProvider) GetForgotPasswordSuccess(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var state *State
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		if anyError == nil && state != nil {
			anyError = state.Error
		}
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUIForgotPasswordSuccessHTML, anyError)
	}

	state, err = p.StateProvider.RestoreState(r, true)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *ForgotPasswordProvider) GetResetPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var state *State
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		if anyError == nil && state != nil {
			anyError = state.Error
		}
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUIResetPasswordHTML, anyError)
	}

	state, err = p.StateProvider.RestoreState(r, true)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *ForgotPasswordProvider) PostResetPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		p.StateProvider.CreateState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			// Remove code from URL
			u := r.URL
			q := u.Query()
			q.Del("code")
			u.RawQuery = q.Encode()
			r.URL = u

			RedirectToPathWithX(w, r, "/reset_password/success")
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate(WebAppSchemaIDResetPasswordRequest, r.Form)
	if err != nil {
		return
	}

	code := r.Form.Get("code")
	newPassword := r.Form.Get("x_password")
	r.Form.Del("x_password")

	err = p.ForgotPassword.ResetPassword(code, newPassword)
	if err != nil {
		return
	}

	return
}

func (p *ForgotPasswordProvider) GetResetPasswordSuccess(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	var state *State
	writeResponse = func(err error) {
		var anyError interface{}
		anyError = err
		if anyError == nil && state != nil {
			anyError = state.Error
		}
		p.RenderProvider.WritePage(w, r, TemplateItemTypeAuthUIResetPasswordSuccessHTML, anyError)
	}

	state, err = p.StateProvider.RestoreState(r, true)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *ForgotPasswordProvider) SetLoginID(r *http.Request) (err error) {
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
