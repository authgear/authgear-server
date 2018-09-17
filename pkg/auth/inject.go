package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"

	"github.com/skygeario/skygear-server/pkg/core/db"
)

type DependencyMap struct{}

func NewDependencyMap() DependencyMap {
	return DependencyMap{}
}

func (m DependencyMap) Provide(dependencyName string, ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "TokenStore":
		return authtoken.NewJWTStore(tConfig.AppName, tConfig.TokenStore.Secret, tConfig.TokenStore.Expiry)
	case "AuthInfoStore":
		return pq.NewAuthInfoStore(
			tConfig.AppName,
			db.NewSQLExecutor(ctx, "postgres", tConfig.DBConnectionStr),
			nil,
		)
	default:
		return nil
	}
}
