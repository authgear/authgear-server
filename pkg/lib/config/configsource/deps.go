package configsource

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewLocalFSLogger,
	wire.Struct(new(LocalFS), "*"),
	NewDatabaseLogger,
	wire.Struct(new(Database), "*"),
	wire.Struct(new(Store), "*"),
)

var ControllerDependencySet = wire.NewSet(
	NewController,
)
