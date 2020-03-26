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

func provideRootHandler(authenticateProvider webapp.AuthenticateProvider) http.Handler {
	return &RootHandler{
		AuthenticateProvider: authenticateProvider,
	}
}

func newRootHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authn.ProvideAuthUIProvider,
		wire.Bind(new(webapp.AuthnProvider), new(*authn.Provider)),
		provideRootHandler,
	)
	return nil
}

func provideSettingsHandler(renderProvider webapp.RenderProvider) http.Handler {
	return &SettingsHandler{RenderProvider: renderProvider}
}

func newSettingsHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideSettingsHandler,
	)
	return nil
}

func provideLogoutHandler(renderProvider webapp.RenderProvider, sm logoutSessionManager) http.Handler {
	return &LogoutHandler{
		RenderProvider: renderProvider,
		SessionManager: sm,
	}
}

func newLogoutHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		wire.Bind(new(logoutSessionManager), new(*auth.SessionManager)),
		provideLogoutHandler,
	)
	return nil
}
