//go:build wireinject
// +build wireinject

package images

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/images/handler"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newPanicMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.PanicMiddleware)),
	))
}

func newSentryMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		wire.Struct(new(middleware.SentryMiddleware), "*"),
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newGetHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.DependencySet,
		handler.DependencySet,
		wire.Bind(new(http.Handler), new(*handler.GetHandler)),
	))
}
