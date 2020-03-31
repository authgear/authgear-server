package gateway

import (
	"context"
	"net/http"

	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
)

type DependencyMap struct {
	Config gatewayConfig.Configuration
}

// nolint: golint
func (m DependencyMap) Provide(
	dependencyName string,
	request *http.Request,
	ctx context.Context,
	requestID string,
	tConfig config.TenantConfiguration,
) interface{} {
	newLoggerFactory := func() logging.Factory {
		logHook := logging.NewDefaultLogHook(tConfig.DefaultSensitiveLoggerValues())
		sentryHook := sentry.NewLogHookFromContext(ctx)
		if request == nil {
			return logging.NewFactoryFromRequestID(requestID, logHook, sentryHook)
		} else {
			return logging.NewFactoryFromRequest(request, logHook, sentryHook)
		}
	}

	switch dependencyName {
	case "LoggerFactory":
		return newLoggerFactory()
	case "AuthInfoStore":
		return pqAuthInfo.NewAuthInfoStore(
			db.NewSQLBuilder("core", tConfig.DatabaseConfig.DatabaseSchema, tConfig.AppID),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
		)
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	case "GatewayConfiguration":
		return m.Config
	default:
		return nil
	}
}
