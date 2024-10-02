package ratelimit

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Limiter), "*"),
	wire.Bind(new(Storage), new(*StorageRedis)),
	NewAppStorageRedis,
)
