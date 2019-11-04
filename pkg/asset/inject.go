package asset

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/asset/dependency/presign"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	redisSession "github.com/skygeario/skygear-server/pkg/core/auth/session/redis"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type DependencyMap struct {
	Storage           cloudstorage.Storage
	Validator         *validation.Validator
	UseInsecureCookie bool
}

var _ inject.DependencyMap = &DependencyMap{}

// nolint: golint
func (m *DependencyMap) Provide(
	dependencyName string,
	request *http.Request,
	ctx context.Context,
	requestID string,
	tConfig coreConfig.TenantConfiguration,
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

	newAuthContext := func() coreAuth.ContextGetter {
		return coreAuth.NewContextGetterWithContext(ctx)
	}

	newSQLExecutor := func() db.SQLExecutor {
		return db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig))
	}

	newTimeProvider := func() time.Provider {
		return time.NewProvider()
	}

	newSessionProvider := func() session.Provider {
		return session.NewProvider(
			request,
			redisSession.NewStore(ctx, tConfig.AppID, newTimeProvider(), newLoggerFactory()),
			redisSession.NewEventStore(ctx, tConfig.AppID),
			newAuthContext(),
			tConfig.UserConfig.Clients,
		)
	}

	newSessionWriter := func() session.Writer {
		return session.NewWriter(
			newAuthContext(),
			tConfig.UserConfig.Clients,
			tConfig.UserConfig.MFA,
			m.UseInsecureCookie,
		)
	}

	newAuthInfoStore := func() authinfo.Store {
		return pqAuthInfo.NewAuthInfoStore(
			db.NewSQLBuilder("core", tConfig.AppConfig.DatabaseSchema, tConfig.AppID),
			newSQLExecutor(),
		)
	}

	switch dependencyName {
	case "APIClientConfigurationProvider":
		return apiclientconfig.NewProvider(newAuthContext(), tConfig)
	case "AuthContextGetter":
		return newAuthContext()
	case "AuthContextSetter":
		return coreAuth.NewContextSetterWithContext(ctx)
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	case "LoggerFactory":
		return newLoggerFactory()
	case "RequireAuthz":
		return handler.NewRequireAuthzFactory(newAuthContext(), newLoggerFactory())
	case "SessionProvider":
		return newSessionProvider()
	case "SessionWriter":
		return newSessionWriter()
	case "AuthInfoStore":
		return newAuthInfoStore()
	case "CloudStorageProvider":
		return cloudstorage.NewProvider(
			tConfig.AppID,
			m.Storage,
			tConfig.UserConfig.Asset.Secret,
			newTimeProvider(),
		)
	case "Validator":
		return m.Validator
	case "PresignProvider":
		return presign.NewProvider(tConfig.UserConfig.Asset.Secret, newTimeProvider())
	default:
		return nil
	}
}
