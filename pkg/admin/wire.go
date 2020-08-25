//+build wireinject

package admin

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/admin/transport"
	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newSentryMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newRootRecoverMiddleware(p *deps.RootProvider) httproute.Middleware {
	panic(wire.Build(
		deps.RootDependencySet,
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.RecoverMiddleware)),
	))
}

func newRequestRecoverMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.RecoverMiddleware)),
	))
}

func newAuthorizationMiddleware(p *deps.RequestProvider, auth config.AdminAPIAuth) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(httproute.Middleware), new(*adminauthz.Middleware)),
	))
}

func newGraphQLHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.GraphQLHandler)),
	))
}
