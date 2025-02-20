package resolver

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	middleware.DependencySet,

	handler.DependencySet,
	wire.Bind(new(handler.Database), new(*appdb.Handle)),
	wire.Bind(new(handler.UserProvider), new(*user.Queries)),
	wire.Bind(new(handler.RolesAndGroupsProvider), new(*rolesgroups.Queries)),
)
