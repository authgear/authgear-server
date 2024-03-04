package pgsearch

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewStore,
	wire.Struct(new(Service), "*"),
	NewLogger,
	wire.Struct(new(Sink), "*"),
)
