package global

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewHandle,
	NewSQLExecutor,
	NewSQLBuilder,
)
