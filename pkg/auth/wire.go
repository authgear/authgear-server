//+build wireinject

package auth

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type middlewareInstance interface {
	Handle(next http.Handler) http.Handler
}

func provideMiddleware(m middlewareInstance) mux.MiddlewareFunc {
	return m.Handle
}

func NewAccessKeyMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(
		DependencySet,
		provideMiddleware,
		wire.Bind(new(middlewareInstance), new(*auth.AccessKeyMiddleware)),
	)
	return nil
}

func NewSessionMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(
		DependencySet,
		provideMiddleware,
		wire.Bind(new(middlewareInstance), new(*session.Middleware)),
	)
	return nil
}

func NewCSPMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(DependencySet, webapp.ProvideCSPMiddleware)
	return nil
}
