package deps

import "github.com/google/wire"

var commonDependencySet = wire.NewSet(
	wire.FieldsOf(new(RootContainer),
		"ServerConfig",
		"DatabasePool",
		"RedisPool",
		"AsyncTaskExecutor",
		"ReservedNameChecker",
	),
)

var RootDependencySet = wire.NewSet(
	commonDependencySet,
	wire.FieldsOf(new(RootContainer),
		"LoggerFactory",
	),
)

var RequestDependencySet = wire.NewSet(
	commonDependencySet,
	wire.FieldsOf(new(RequestContainer),
		"RootContainer",
		"LoggerFactory",
	),
)
