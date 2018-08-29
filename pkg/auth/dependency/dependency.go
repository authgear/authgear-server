package dependency

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type AuthDependency struct {
	dbProvider *db.DBProvider
}

// SetDBProvider set a DB provider implementation to dependency graph
func (d *AuthDependency) SetDBProvider(dbProvider db.DBProvider) {
	d.dbProvider = &dbProvider
}

func (d AuthDependency) Provide(dependencyName string, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return func() db.IDB {
			dbProvider := d.dbProvider
			if dbProvider == nil {
				return nil
			}

			return (*dbProvider).GetDB(tConfig)
		}
	default:
		return nil
	}
}
