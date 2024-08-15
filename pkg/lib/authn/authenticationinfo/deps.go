package authenticationinfo

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StoreRedis), "*"),
	wire.Struct(new(UIService), "*"),
)
