package executor

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewInProcessExecutorLogger,
	wire.Struct(new(InProcessExecutor), "*"),
)
