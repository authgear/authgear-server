//+build wireinject

package server

import (
	getsentry "github.com/getsentry/sentry-go"
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/core/sentry"
	"github.com/authgear/authgear-server/pkg/deps"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/middlewares"
)

var rootMiddlewareDependencySet = wire.NewSet(
	deps.RootDependencySet,
)

var middlewareDependencySet = wire.NewSet(
	deps.RequestDependencySet,
)

func newSentryMiddlewareFactory(hub *getsentry.Hub) func(*deps.RootProvider) httproute.Middleware {
	return func(p *deps.RootProvider) httproute.Middleware {
		return newSentryMiddleware(hub, p)
	}
}

func newSentryMiddleware(hub *getsentry.Hub, p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		rootMiddlewareDependencySet,
		sentry.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*sentry.Middleware)),
	))
}

func newRecoverMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		rootMiddlewareDependencySet,
		middlewares.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middlewares.RecoverMiddleware)),
	))
}

func newCORSMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*middlewares.CORSMiddleware)),
	))
}

func newCSPMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.CSPMiddleware)),
	))
}

func newCSRFMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.CSRFMiddleware)),
	))
}

func newAuthEntryPointMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.AuthEntryPointMiddleware)),
	))
}

func newSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*auth.Middleware)),
	))
}

func newWebAppStateMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*webapp.StateMiddleware)),
	))
}
