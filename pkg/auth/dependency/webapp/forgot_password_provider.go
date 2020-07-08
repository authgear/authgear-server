package webapp

import ()

// func (p *ForgotPasswordProvider) PostResetPasswordForm(w http.ResponseWriter, r *http.Request) (writeResponse func(err error), err error) {
// 	writeResponse = func(err error) {
// 		p.StateProvider.CreateState(r, nil, err)
// 		if err != nil {
// 			RedirectToCurrentPath(w, r)
// 		} else {
// 			// Remove code from URL
// 			u := r.URL
// 			q := u.Query()
// 			q.Del("code")
// 			u.RawQuery = q.Encode()
// 			r.URL = u
//
// 			RedirectToPathWithX(w, r, "/reset_password/success")
// 		}
// 	}
//
// 	p.ValidateProvider.PrepareValues(r.Form)
//
// 	err = p.ValidateProvider.Validate(WebAppSchemaIDResetPasswordRequest, r.Form)
// 	if err != nil {
// 		return
// 	}
//
// 	code := r.Form.Get("code")
// 	newPassword := r.Form.Get("x_password")
// 	r.Form.Del("x_password")
//
// 	err = p.ForgotPassword.ResetPassword(code, newPassword)
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
