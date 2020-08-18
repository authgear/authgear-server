package admin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/admin/transport"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	middleware.DependencySet,

	loader.DependencySet,
	wire.Bind(new(loader.UserService), new(*user.Queries)),

	graphql.DependencySet,
	wire.Bind(new(graphql.UserLoader), new(*loader.UserLoader)),

	transport.DependencySet,
)
