package provider

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type AuthProviders struct {
	DB            *db.DBProvider
	TokenStore    *auth.TokenStoreProvider
	AuthInfoStore *auth.AuthInfoStoreProvider
}

func (d AuthProviders) Provide(dependencyName string, ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return d.DB.Provide(ctx, tConfig)
	case "TokenResolver":
		return d.TokenStore.Provide(ctx, tConfig)
	case "AuthInfoStore":
		return d.AuthInfoStore.Provide(ctx, tConfig)
	default:
		return nil
	}
}
