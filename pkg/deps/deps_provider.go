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

var commonDeps = wire.NewSet(
	configDeps,
)

var RequestDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
	),
	commonDeps,
)

var TaskDependencySet = wire.NewSet(
	wire.FieldsOf(new(*TaskProvider),
		"AppProvider",
	),
	commonDeps,
)
