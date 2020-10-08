package db

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Handle), "*"),
	NewSQLBuilder,
	wire.Struct(new(SQLExecutor), "*"),
)
