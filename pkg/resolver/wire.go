//go:build wireinject
// +build wireinject

package resolver

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newHealthzHandler(p *deps.RootProvider, w http.ResponseWriter, r *http.Request, ctx context.Context) http.Handler {
	panic(wire.Build(
		deps.RootDependencySet,
		healthz.DependencySet,
		wire.Bind(new(http.Handler), new(*healthz.Handler)),
	))
}

func newPanicMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.PanicMiddleware)),
	))
}

func newSentryMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newBodyLimitMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.BodyLimitMiddleware)),
	))
}

func newSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*session.Middleware)),
	))
}

func newSessionResolveHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*handler.ResolveHandler)),
	))
}
