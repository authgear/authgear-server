//+build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

var rootMiddlewareDependencySet = wire.NewSet(
	deps.RootDependencySet,
)

var middlewareDependencySet = wire.NewSet(
	deps.RequestDependencySet,
)

func newSentryMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		rootMiddlewareDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newRootRecoverMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		rootMiddlewareDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.RecoverMiddleware)),
	))
}

func newRequestRecoverMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.RecoverMiddleware)),
	))
}

func newCORSMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middlewareDependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.CORSMiddleware)),
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
