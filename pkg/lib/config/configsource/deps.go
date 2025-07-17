package configsource

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(LocalFS), "*"),
	NewDatabaseHandleFactory,
	NewConfigSourceStoreStoreFactory,
	NewPlanStoreStoreFactory,
	wire.Struct(new(Database), "*"),
	wire.Struct(new(Store), "*"),
)

var ControllerDependencySet = wire.NewSet(
	NewController,
)
