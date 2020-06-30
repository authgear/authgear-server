package session

import (
	"github.com/google/wire"

	corerand "github.com/authgear/authgear-server/pkg/core/rand"
)

var DependencySet = wire.NewSet(
	NewSessionCookieDef,
	wire.Value(Rand(corerand.SecureRand)),
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(Resolver), "*"),
	wire.Struct(new(Manager), "*"),
	wire.Bind(new(resolverProvider), new(*Provider)),
)
