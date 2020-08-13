package source

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewLocalFileLogger,
	wire.Struct(new(LocalFile), "*"),

	NewSource,
)
