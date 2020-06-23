package session

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	corerand "github.com/skygeario/skygear-server/pkg/core/rand"
)

var DependencySet = wire.NewSet(
	NewSessionCookieDef,
	wire.Value(Rand(corerand.SecureRand)),
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(Resolver), "*"),
	wire.Bind(new(auth.IDPSessionResolver), new(*Resolver)),
	wire.Struct(new(Manager), "*"),
	wire.Bind(new(auth.IDPSessionManager), new(*Manager)),
)
