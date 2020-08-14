package resolver

import (
	"github.com/google/wire"

	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
)

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	// FIXME: refactor these services?
	wire.InterfaceValue(new(verification.WebAppURLProvider), verification.WebAppURLProvider(nil)),
	wire.InterfaceValue(new(otp.EndpointsProvider), otp.EndpointsProvider(nil)),

	middleware.DependencySet,

	handler.DependencySet,
	wire.Bind(new(handler.IdentityService), new(*identityservice.Service)),
	wire.Bind(new(handler.VerificationService), new(*verification.Service)),
)
