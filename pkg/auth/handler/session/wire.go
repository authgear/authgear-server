//+build wireinject

package session

import (
	"net/http"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func provideResolveHandler(
	m *auth.Middleware,
	lf logging.Factory,
	t clock.Clock,
	ap *anonymous.Provider,
) http.Handler {
	return m.Handle(&ResolveHandler{
		LoggerFactory: lf,
		Anonymous:     ap,
	})
}

func newResolveHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(pkg.DependencySet, provideResolveHandler)
	return nil
}
