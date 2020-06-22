package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/deps"
)

func newLoginHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newEnterPasswordHandler(p *deps.RequestProvider) http.Handler {
	return (*EnterPasswordHandler)(nil)
}

func newForgotPasswordHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newForgotPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newResetPasswordHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newResetPasswordSuccessHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newSignupHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newPromoteHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newCreatePasswordHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newSettingsHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newSettingsIdentityHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newOOBOTPHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newEnterLoginIDHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newLogoutHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}

func newSSOCallbackHandler(p *deps.RequestProvider) http.Handler {
	return (*LoginHandler)(nil)
}
