//+build wireinject

package webapp

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

var authDepSet = wire.NewSet(
	authn.ProvideAuthUIProvider,
	wire.Bind(new(webapp.AuthnProvider), new(*authn.Provider)),
	wire.Struct(new(webapp.AuthenticateProviderImpl), "*"),
)

func newLoginHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authDepSet,
		wire.Bind(new(loginProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(LoginHandler), "*"),
		wire.Bind(new(http.Handler), new(*LoginHandler)),
	)
	return nil
}

func newLoginPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authDepSet,
		wire.Bind(new(loginPasswordProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(LoginPasswordHandler), "*"),
		wire.Bind(new(http.Handler), new(*LoginPasswordHandler)),
	)
	return nil
}

func newSignupHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authDepSet,
		wire.Bind(new(signupProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(SignupHandler), "*"),
		wire.Bind(new(http.Handler), new(*SignupHandler)),
	)
	return nil
}

func newSignupPasswordHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authDepSet,
		wire.Bind(new(signupPasswordProvider), new(*webapp.AuthenticateProviderImpl)),
		wire.Struct(new(SignupPasswordHandler), "*"),
		wire.Bind(new(http.Handler), new(*SignupPasswordHandler)),
	)
	return nil
}

func newSettingsHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Struct(new(SettingsHandler), "*"),
		wire.Bind(new(http.Handler), new(*SettingsHandler)),
	)
	return nil
}

func newLogoutHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(logoutSessionManager), new(*auth.SessionManager)),
		wire.Struct(new(LogoutHandler), "*"),
		wire.Bind(new(http.Handler), new(*LogoutHandler)),
	)
	return nil
}
