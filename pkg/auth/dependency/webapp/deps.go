package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StateStoreImpl), "*"),
	wire.Bind(new(StateStore), new(*StateStoreImpl)),
	NewStateProviderLogger,
	wire.Struct(new(StateProviderImpl), "*"),
	wire.Bind(new(StateProvider), new(*StateProviderImpl)),
	wire.Struct(new(URLProvider), "*"),
	wire.Struct(new(OAuthService), "*"),
	wire.Struct(new(Responder), "*"),

	wire.Struct(new(CSPMiddleware), "*"),
	wire.Struct(new(CSRFMiddleware), "*"),
	wire.Struct(new(AuthEntryPointMiddleware), "*"),
	wire.Struct(new(StateMiddleware), "*"),
)
