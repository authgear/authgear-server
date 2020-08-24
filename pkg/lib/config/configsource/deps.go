package configsource

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewLocalFSLogger,
	wire.Struct(new(LocalFS), "*"),

	NewController,
)
