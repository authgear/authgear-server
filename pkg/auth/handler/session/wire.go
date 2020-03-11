//+build wireinject

package session

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func provideResolveHandler(m *session.Middleware, t time.Provider) http.Handler {
	return m.Handle(&ResolveHandler{
		TimeProvider: t,
	})
}

func newResolveHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(auth.DependencySet, provideResolveHandler)
	return nil
}
