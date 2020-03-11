//+build wireinject

package session

import (
	"net/http"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
)

func provideResolveHandler(m *session.Middleware) http.Handler {
	return nil
}

func newResolveHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(auth.DependencySet, provideResolveHandler)
	return nil
}
