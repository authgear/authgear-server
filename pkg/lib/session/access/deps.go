package access

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(EventStoreRedis), "*"),
	wire.Bind(new(EventStore), new(*EventStoreRedis)),
	wire.Struct(new(EventProvider), "*"),
)
