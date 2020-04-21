package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type ForgotPassword interface {
	SendCode(loginID string) error
	ResetPassword(code string, newPassword string) error
}

type ForgotPasswordProvider struct {
	ValidateProvider ValidateProvider
	RenderProvider   RenderProvider
	StateStore       StateStore
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

	state, err = p.restoreState(r)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *ForgotPasswordProvider) PostForgotPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			RedirectToPathWithQueryPreserved(w, r, "/forgot_password/success")
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppForgotPasswordRequest", r.Form)
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

	state, err = p.restoreState(r)
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

	state, err = p.restoreState(r)
	if err != nil {
		return
	}

	p.ValidateProvider.PrepareValues(r.Form)

	return
}

func (p *ForgotPasswordProvider) PostResetPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
	writeResponse = func(err error) {
		p.persistState(r, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			// Remove code from URL
			u := r.URL
			q := u.Query()
			q.Del("code")
			u.RawQuery = q.Encode()
			r.URL = u

			RedirectToPathWithQueryPreserved(w, r, "/reset_password/success")
		}
	}

	p.ValidateProvider.PrepareValues(r.Form)

	err = p.ValidateProvider.Validate("#WebAppResetPasswordRequest", r.Form)
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

	state, err = p.restoreState(r)
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

func (p *ForgotPasswordProvider) persistState(r *http.Request, inputError error) {
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

func (p *ForgotPasswordProvider) restoreState(r *http.Request) (state *State, err error) {
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
