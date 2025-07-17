package lockout

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StorageRedis), "*"),
	wire.Struct(new(Service), "*"),
	wire.Bind(new(Storage), new(*StorageRedis)),
)
