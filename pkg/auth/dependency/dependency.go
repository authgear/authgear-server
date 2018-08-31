package dependency

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type AuthDependency struct {
	DB *db.DBProvider
}

func (d AuthDependency) Provide(dependencyName string, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return d.DB.Provide(tConfig)
	default:
		return nil
	}
}
