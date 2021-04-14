package db

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Handle), "*"),
	NewSQLBuilder,
	wire.Struct(new(SQLExecutor), "*"),
)

func NewSQLBuilder(config *config.DatabaseEnvironmentConfig) *db.SQLBuilder {
	builder := db.NewSQLBuilder(config.DatabaseSchema, "")
	return &builder
}
