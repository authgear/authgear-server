package provider

import (
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type AuthProviders struct {
	DB         *db.DBProvider
	TokenStore *auth.TokenStoreProvider
}

func (d AuthProviders) Provide(dependencyName string, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return d.DB.Provide(tConfig)
	case "TokenResolver":
		return d.TokenStore.Provide(tConfig)
	default:
		return nil
	}
}
