package auth

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	pqRole "github.com/skygeario/skygear-server/pkg/core/auth/role/pq"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/logging"
)

func openDB(tConfig config.TenantConfiguration) func() (*sqlx.DB, error) {
	return func() (*sqlx.DB, error) {
		return sqlx.Open("postgres", tConfig.DBConnectionStr)
	}
}

func NewDefaultTokenStore(ctx context.Context, tConfig config.TenantConfiguration) authtoken.Store {
	return authtoken.NewJWTStore(tConfig.AppName, tConfig.TokenStore.Secret, tConfig.TokenStore.Expiry)
}

func NewDefaultAuthInfoStore(ctx context.Context, tConfig config.TenantConfiguration) authinfo.Store {
	return pqAuthInfo.NewSafeAuthInfoStore(
		db.NewSQLBuilder("core", tConfig.AppName),
		db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, openDB(tConfig))),
		logging.CreateLogger(ctx, "authinfo"),
		db.NewSafeTxContextWithContext(ctx, openDB(tConfig)),
	)
}

func NewDefaultRoleStore(ctx context.Context, tConfig config.TenantConfiguration) role.Store {
	return pqRole.NewSafeRoleStore(
		db.NewSQLBuilder("core", tConfig.AppName),
		db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, openDB(tConfig))),
		logging.CreateLogger(ctx, "role"),
		db.NewSafeTxContextWithContext(ctx, openDB(tConfig)),
	)
}
