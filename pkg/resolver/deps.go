package resolver

import (
	"github.com/google/wire"

	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	middleware.DependencySet,

	handler.DependencySet,
	wire.Bind(new(handler.Database), new(*tenantdb.Handle)),
	wire.Bind(new(handler.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(handler.VerificationService), new(*verification.Service)),

	wire.Struct(new(EndpointsProvider), "*"),
	wire.Bind(new(oauth.BaseURLProvider), new(*EndpointsProvider)),
	wire.Bind(new(oidc.BaseURLProvider), new(*EndpointsProvider)),
)
