package db

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	globaldb.NewHandle,
	NewSQLBuilder,
	db.NewGlobalSQLExecutor,
)

func NewSQLBuilder(config *config.DatabaseEnvironmentConfig) *db.SQLBuilder {
	builder := db.NewSQLBuilder(config.DatabaseSchema, "")
	return &builder
}
