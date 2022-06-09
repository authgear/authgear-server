package meter

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewStoreRedisLogger,
	wire.Struct(new(Service), "*"),
	wire.Struct(new(ReadStoreRedis), "*"),
	wire.Struct(new(WriteStoreRedis), "*"),
	wire.Bind(new(CounterStore), new(*WriteStoreRedis)),
)
