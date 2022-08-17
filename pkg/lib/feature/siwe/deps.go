package siwe

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(StoreRedis), "*"),
	wire.Bind(new(NonceStore), new(*StoreRedis)),
	NewLogger,
	wire.Struct(new(Service), "*"),
)
