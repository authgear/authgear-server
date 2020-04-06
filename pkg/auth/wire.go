//+build wireinject

package auth

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
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
		wire.Bind(new(middlewareInstance), new(*coreauth.AccessKeyMiddleware)),
	)
	return nil
}

func NewSessionMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(
		DependencySet,
		provideMiddleware,
		wire.Bind(new(middlewareInstance), new(*auth.Middleware)),
	)
	return nil
}

func NewCSPMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(DependencySet, webapp.ProvideCSPMiddleware)
	return nil
}

func NewCSRFMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(DependencySet, ProvideCSRFMiddleware)
	return nil
}

func NewStateMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	wire.Build(DependencySet, webapp.ProvideStateMiddleware)
	return nil
}

func newSessionManager(r *http.Request, m DependencyMap) *auth.SessionManager {
	wire.Build(DependencySet)
	return nil
}
