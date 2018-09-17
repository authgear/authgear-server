package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/db"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreDB "github.com/skygeario/skygear-server/pkg/core/db"
)

type DependencyMap struct {
	DB *coreDB.DBProvider
}

func NewDependencyMap() DependencyMap {
	return DependencyMap{
		DB: coreDB.NewDBProvider("auth"),
	}
}

func (m DependencyMap) Provide(dependencyName string, ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return m.ProvideDB(ctx, tConfig)
	case "TokenStore":
		return authtoken.NewJWTStore(tConfig.AppName, tConfig.TokenStore.Secret, tConfig.TokenStore.Expiry)
	case "AuthInfoStore":
		return pq.AuthInfoStore{}
	default:
		return nil
	}
}

func (m DependencyMap) ProvideDB(ctx context.Context, tConfig config.TenantConfiguration) *db.DBConn {
	conn := m.DB.Provide(ctx, tConfig)
	db := &db.DBConn{
		conn,
		tConfig.DBConnectionStr,
	}
	return db
}
