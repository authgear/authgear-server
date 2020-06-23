package deps

import "github.com/google/wire"

var taskDeps = wire.NewSet(
	commonDeps,
)
