//go:build wireinject
// +build wireinject

package superadmin

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newPanicMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.PanicMiddleware)),
	))
}

func newBodyLimitMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.BodyLimitMiddleware)),
	))
}

func newOtelMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		otelauthgear.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*otelauthgear.HTTPInstrumentationMiddleware)),
	))
}

func newSentryMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		wire.Struct(new(middleware.SentryMiddleware), "*"),
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newSessionInfoMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		session.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*session.SessionInfoMiddleware)),
	))
}

func newHealthzHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		healthz.DependencySet,
		wire.Bind(new(http.Handler), new(*healthz.Handler)),
	))
}

func newGraphQLHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.GraphQLHandler)),
	))
}
