//+build wireinject

package main

import (
	"net/http"

	getsentry "github.com/getsentry/sentry-go"
	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/middlewares"
)

type middleware interface {
	Handle(next http.Handler) http.Handler
}

func provideMiddlewareFunc(m middleware) mux.MiddlewareFunc { return m.Handle }

var rootMiddlewareDependencySet = wire.NewSet(
	deps.RootDependencySet,
	provideMiddlewareFunc,
)

var middlewareDependencySet = wire.NewSet(
	deps.RequestDependencySet,
	provideMiddlewareFunc,
)

func newSentryMiddlewareFactory(hub *getsentry.Hub) func(*deps.RootProvider) mux.MiddlewareFunc {
	return func(p *deps.RootProvider) mux.MiddlewareFunc {
		return newSentryMiddleware(hub, p)
	}
}

func newSentryMiddleware(hub *getsentry.Hub, p *deps.RootProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		rootMiddlewareDependencySet,
		sentry.DependencySet,
		wire.Bind(new(middleware), new(*sentry.Middleware)),
	))
}

func newRecoverMiddleware(p *deps.RootProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		rootMiddlewareDependencySet,
		middlewares.DependencySet,
		wire.Bind(new(middleware), new(*middlewares.RecoverMiddleware)),
	))
}

func newCORSMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(middleware), new(*middlewares.CORSMiddleware)),
	))
}

func newCSPMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(middleware), new(*webapp.CSPMiddleware)),
	))
}

func newCSRFMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(middleware), new(*webapp.CSRFMiddleware)),
	))
}

func newAuthEntryPointMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(middleware), new(*webapp.AuthEntryPointMiddleware)),
	))
}

func newSessionMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(middleware), new(*auth.Middleware)),
	))
}

func newWebAppStateMiddleware(p *deps.RequestProvider) mux.MiddlewareFunc {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(middleware), new(*webapp.StateMiddleware)),
	))
}
