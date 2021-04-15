package tenant

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func ProvideSQLBuilder(c *config.DatabaseCredentials, id config.AppID) db.SQLBuilder {
	return db.NewSQLBuilder(c.DatabaseSchema, string(id))
}

var DependencySet = wire.NewSet(
	NewSQLExecutor,
	ProvideSQLBuilder,
)
