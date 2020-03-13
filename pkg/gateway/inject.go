package gateway

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	redisSession "github.com/skygeario/skygear-server/pkg/core/auth/session/redis"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/core/time"
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
	newAuthContext := func() auth.ContextGetter {
		return auth.NewContextGetterWithContext(ctx)
	}

	switch dependencyName {
	case "AuthContextGetter":
		return newAuthContext()
	case "AuthContextSetter":
		return auth.NewContextSetterWithContext(ctx)
	case "LoggerFactory":
		return newLoggerFactory()
	case "SessionProvider":
		return session.NewProvider(
			request,
			redisSession.NewStore(ctx, tConfig.AppID, time.NewProvider(), newLoggerFactory()),
			redisSession.NewEventStore(ctx, tConfig.AppID),
			tConfig.AppConfig.Clients,
		)
	case "SessionWriter":
		return session.NewWriter(
			ctx,
			tConfig.AppConfig.Clients,
			tConfig.AppConfig.MFA,
			m.Config.UseInsecureCookie,
		)
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
