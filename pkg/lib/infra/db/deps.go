package db

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func ProvideSQLBuilder(c *config.DatabaseCredentials, id config.AppID) SQLBuilder {
	return NewSQLBuilder(c.DatabaseSchema, string(id))
}

var DependencySet = wire.NewSet(
	NewTenantSQLExecutor,
	ProvideSQLBuilder,
)
