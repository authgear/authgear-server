//go:build wireinject
// +build wireinject

package portal

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newPanicMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
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

func newGraphQLHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.GraphQLHandler)),
	))
}

func newSystemConfigHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.SystemConfigHandler)),
	))
}

func newAdminAPIHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.AdminAPIHandler)),
	))
}

func newStaticAssetsHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.StaticAssetsHandler)),
	))
}

func newStripeWebhookHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.StripeWebhookHandler)),
	))
}

func newOsanoHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.OsanoHandler)),
	))
}
