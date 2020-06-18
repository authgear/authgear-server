package deps

import "github.com/google/wire"

var commonDependencySet = wire.NewSet(
	wire.FieldsOf(new(RootProvider),
		"ServerConfig",
		"DatabasePool",
		"RedisPool",
		"AsyncTaskExecutor",
		"ReservedNameChecker",
	),
)

var RootDependencySet = wire.NewSet(
	commonDependencySet,
	wire.FieldsOf(new(RootProvider),
		"LoggerFactory",
	),
)

var RequestDependencySet = wire.NewSet(
	commonDependencySet,
	wire.FieldsOf(new(RequestProvider),
		"RootProvider",
		"LoggerFactory",
	),
)
