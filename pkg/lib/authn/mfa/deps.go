package mfa

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewDeviceTokenCookieDef,
	wire.Struct(new(StoreDeviceTokenRedis), "*"),
	wire.Bind(new(StoreDeviceToken), new(*StoreDeviceTokenRedis)),
	wire.Struct(new(StoreRecoveryCodePQ), "*"),
	wire.Bind(new(StoreRecoveryCode), new(*StoreRecoveryCodePQ)),
	wire.Struct(new(ReadOnlyService), "*"),
	wire.Struct(new(Service), "*"),
	wire.Struct(new(Lockout), "*"),
)
