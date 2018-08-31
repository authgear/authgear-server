package provider

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type AuthProviders struct {
	DB *db.DBProvider
}

func (d AuthProviders) Provide(dependencyName string, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return d.DB.Provide(tConfig)
	default:
		return nil
	}
}
