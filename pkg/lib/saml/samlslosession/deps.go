package samlslosession

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StoreRedis), "*"),
)
