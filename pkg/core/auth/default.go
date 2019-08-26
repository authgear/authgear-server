package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func NewDefaultAuthInfoStore(ctx context.Context, tConfig config.TenantConfiguration) authinfo.Store {
	return pqAuthInfo.NewSafeAuthInfoStore(
		db.NewSQLBuilder("core", tConfig.AppConfig.DatabaseSchema, tConfig.AppID),
		db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
		logging.CreateLoggerWithContext(ctx, "authinfo"),
		db.NewSafeTxContextWithContext(ctx, tConfig),
	)
}
