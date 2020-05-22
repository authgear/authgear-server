package asset

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/asset/dependency/presign"
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
	tConfig coreConfig.TenantConfiguration,
) interface{} {
	newLoggerFactory := func() logging.Factory {
		logHook := logging.NewDefaultLogHook(tConfig.DefaultSensitiveLoggerValues())
		sentryHook := sentry.NewLogHookFromContext(ctx)
		return logging.NewFactory(logHook, sentryHook)
	}

	newTimeProvider := func() time.Provider {
		return time.NewProvider()
	}

	switch dependencyName {
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	case "LoggerFactory":
		return newLoggerFactory()
	case "RequireAuthz":
		return handler.NewRequireAuthzFactory(newLoggerFactory())
	case "CloudStorageProvider":
		return cloudstorage.NewProvider(
			tConfig.AppID,
			m.Storage,
			tConfig.AppConfig.Asset.Secret,
			newTimeProvider(),
		)
	case "Validator":
		return m.Validator
	case "PresignProvider":
		return presign.NewProvider(tConfig.AppConfig.Asset.Secret, newTimeProvider())
	default:
		return nil
	}
}
