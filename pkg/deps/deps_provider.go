package deps

import (
	"github.com/google/wire"
)

var appRootDeps = wire.NewSet(
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Context",
		"Config",
		"LoggerFactory",
		"DbContext",
		"RedisContext",
	),
	wire.FieldsOf(new(*RootProvider),
		"ServerConfig",
		"TaskExecutor",
		"ReservedNameChecker",
	),
)

var RequestDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
	),
	requestDeps,
)

var TaskDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*TaskProvider),
		"AppProvider",
	),
	taskDeps,
)
