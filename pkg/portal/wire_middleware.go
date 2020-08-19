//+build wireinject

package portal

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/upstreamapp"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newRecoverMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		middleware.NewRecoveryLogger,
		wire.Struct(new(middleware.RecoverMiddleware), "*"),
		wire.Bind(new(httproute.Middleware), new(*middleware.RecoverMiddleware)),
	))
}

func newSessionInfoMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		upstreamapp.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*upstreamapp.Middleware)),
	))
}
