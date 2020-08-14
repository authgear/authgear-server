package interaction

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Context), "*"),
	wire.Struct(new(StoreRedis), "*"),
	wire.Bind(new(Store), new(*StoreRedis)),
	NewLogger,
	wire.Struct(new(Service), "*"),
)
