//+build wireinject

package webapp

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpassword"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func provideRedirectURIForWebAppFunc() sso.RedirectURLFunc {
	return redirectURIForWebApp
}

var dependencySet = wire.NewSet(
	pkg.DependencySet,
	authn.ProvideAuthUIProvider,
	wire.Bind(new(webapp.AuthnProvider), new(*authn.Provider)),
	wire.Bind(new(webapp.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),
	wire.Struct(new(webapp.AuthenticateProviderImpl), "*"),
	wire.Bind(new(webapp.ForgotPassword), new(*forgotpassword.Provider)),
	wire.Struct(new(webapp.ForgotPasswordProvider), "*"),
	provideRedirectURIForWebAppFunc,
)

func newLoginHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(loginProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(LoginHandler), "*"),
		wire.Bind(new(http.Handler), new(*LoginHandler)),
	)
	return nil
}

func newEnterPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(enterPasswordProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(EnterPasswordHandler), "*"),
		wire.Bind(new(http.Handler), new(*EnterPasswordHandler)),
	)
	return nil
}

func newForgotPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(forgotPasswordProvider), new(*webapp.ForgotPasswordProvider)),
		wire.Struct(new(ForgotPasswordHandler), "*"),
		wire.Bind(new(http.Handler), new(*ForgotPasswordHandler)),
	)
	return nil
}

func newForgotPasswordSuccessHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(forgotPasswordSuccessProvider), new(*webapp.ForgotPasswordProvider)),
		wire.Struct(new(ForgotPasswordSuccessHandler), "*"),
		wire.Bind(new(http.Handler), new(*ForgotPasswordSuccessHandler)),
	)
	return nil
}

func newResetPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(resetPasswordProvider), new(*webapp.ForgotPasswordProvider)),
		wire.Struct(new(ResetPasswordHandler), "*"),
		wire.Bind(new(http.Handler), new(*ResetPasswordHandler)),
	)
	return nil
}

func newResetPasswordSuccessHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(resetPasswordSuccessProvider), new(*webapp.ForgotPasswordProvider)),
		wire.Struct(new(ResetPasswordSuccessHandler), "*"),
		wire.Bind(new(http.Handler), new(*ResetPasswordSuccessHandler)),
	)
	return nil
}

func newSignupHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(signupProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(SignupHandler), "*"),
		wire.Bind(new(http.Handler), new(*SignupHandler)),
	)
	return nil
}

func newPromoteHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(promoteProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(PromoteHandler), "*"),
		wire.Bind(new(http.Handler), new(*PromoteHandler)),
	)
	return nil
}

func newCreatePasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(createPasswordProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(CreatePasswordHandler), "*"),
		wire.Bind(new(http.Handler), new(*CreatePasswordHandler)),
	)
	return nil
}

func newSettingsHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Struct(new(SettingsHandler), "*"),
		wire.Bind(new(http.Handler), new(*SettingsHandler)),
	)
	return nil
}

func newSettingsIdentityHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(settingsIdentityProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(SettingsIdentityHandler), "*"),
		wire.Bind(new(http.Handler), new(*SettingsIdentityHandler)),
	)
	return nil
}

func newOOBOTPHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(OOBOTPProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(OOBOTPHandler), "*"),
		wire.Bind(new(http.Handler), new(*OOBOTPHandler)),
	)
	return nil
}

func newEnterLoginIDHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(EnterLoginIDProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(EnterLoginIDHandler), "*"),
		wire.Bind(new(http.Handler), new(*EnterLoginIDHandler)),
	)
	return nil
}

func newLogoutHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(logoutSessionManager), new(*auth.SessionManager)),
		wire.Struct(new(LogoutHandler), "*"),
		wire.Bind(new(http.Handler), new(*LogoutHandler)),
	)
	return nil
}

func newSSOCallbackHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		dependencySet,
		wire.Bind(new(ssoProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(SSOCallbackHandler), "*"),
		wire.Bind(new(http.Handler), new(*SSOCallbackHandler)),
	)
	return nil
}
