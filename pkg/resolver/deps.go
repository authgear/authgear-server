package resolver

import (
	"github.com/google/wire"

	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	middleware.DependencySet,

	handler.DependencySet,
	wire.Bind(new(handler.Database), new(*appdb.Handle)),
	wire.Bind(new(handler.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(handler.VerificationService), new(*verification.Service)),
	wire.Bind(new(handler.UserProvider), new(*user.Queries)),
	wire.Bind(new(handler.RolesAndGroupsProvider), new(*rolesgroups.Queries)),

	wire.NewSet(
		endpoints.DependencySet,
		wire.Bind(new(oauth.BaseURLProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(oidc.BaseURLProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(tester.EndpointsProvider), new(*endpoints.Endpoints)),
	),

	wire.Bind(new(oauth.OAuthClientResolver), new(*oauthclient.Resolver)),
)
