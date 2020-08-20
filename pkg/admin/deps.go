package admin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/admin/transport"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
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
	wire.Bind(new(loader.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(loader.AuthenticatorService), new(*authenticatorservice.Service)),

	graphql.DependencySet,
	wire.Bind(new(graphql.UserLoader), new(*loader.UserLoader)),
	wire.Bind(new(graphql.IdentityLoader), new(*loader.IdentityLoader)),
	wire.Bind(new(graphql.AuthenticatorLoader), new(*loader.AuthenticatorLoader)),

	transport.DependencySet,
)
