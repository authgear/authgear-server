//+build wireinject

package middlewares

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/deps"
)

type Middleware interface {
	Handle(next http.Handler) http.Handler
}

func provideMiddleware(m Middleware) mux.MiddlewareFunc { return m.Handle }

var depSet = wire.NewSet(
	deps.RequestDependencySet,
	provideMiddleware,
)

func NewSessionMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*auth.Middleware)(nil).Handle
}

func NewCSPMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.CSPMiddleware)(nil).Handle
}

func NewCSRFMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	wire.Build(
		depSet,
		wire.Struct(new(webapp.CSRFMiddleware), "*"),
		wire.Bind(new(Middleware), new(*webapp.CSRFMiddleware)),
	)
	return nil
}

func NewStateMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.StateMiddleware)(nil).Handle
}

func NewAuthEntryPointMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*webapp.AuthEntryPointMiddleware)(nil).Handle
}

func NewCORSMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*CORSMiddleware)(nil).Handle
}

func NewRecoverMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	return (*RecoverMiddleware)(nil).Handle
}
