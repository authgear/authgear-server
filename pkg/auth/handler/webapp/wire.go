//+build wireinject

package webapp

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func provideRootHandler(authenticateProvider webapp.AuthenticateProvider) http.Handler {
	return &RootHandler{
		AuthenticateProvider: authenticateProvider,
	}
}

func newRootHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(webapp.AuthnProvider), new(*authn.Provider)),
		provideRootHandler,
	)
	return nil
}

func provideSettingsHandler(renderProvider webapp.RenderProvider) http.Handler {
	return &SettingsHandler{RenderProvider: renderProvider}
}

func newSettingsHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		provideSettingsHandler,
	)
	return nil
}
