package whatsapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(StoreRedis), "*"),
	NewLogger,
)
