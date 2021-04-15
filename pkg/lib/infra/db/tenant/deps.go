package tenant

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewSQLExecutor,
	NewSQLBuilder,
)
