package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ProvideRequestContext(r *http.Request) context.Context { return r.Context() }

func ProvideConfigSource(ctrl *configsource.Controller) *configsource.ConfigSource {
	return ctrl.GetConfigSource()
}

func ProvideAuditDatabaseCredentials(cfg *config.EnvironmentConfig) *config.AuditDatabaseCredentials {
	if cfg.AuditDatabase.DatabaseURL != "" && cfg.AuditDatabase.DatabaseSchema != "" {
		return &config.AuditDatabaseCredentials{
			DatabaseURL:    cfg.AuditDatabase.DatabaseURL,
			DatabaseSchema: cfg.AuditDatabase.DatabaseSchema,
		}
	}
	return nil
}

func ProvideDatabaseConfig(databaseConfig *config.DatabaseEnvironmentConfig) *config.DatabaseConfig {
	newDurationSecondsConfig := func(sec int) *config.DurationSeconds {
		c := config.DurationSeconds(sec)
		return &c
	}

	return &config.DatabaseConfig{
		MaxOpenConnection:     &databaseConfig.MaxOpenConn,
		MaxIdleConnection:     &databaseConfig.MaxIdleConn,
		IdleConnectionTimeout: newDurationSecondsConfig(databaseConfig.ConnMaxIdleTimeSeconds),
		MaxConnectionLifetime: newDurationSecondsConfig(databaseConfig.ConnMaxLifetimeSeconds),
	}
}

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"AuthgearConfig",
		"AdminAPIConfig",
		"AppConfig",
		"DatabaseConfig",
		"SMTPConfig",
		"MailConfig",
		"KubernetesConfig",
		"DomainImplementation",
		"SearchConfig",
		"AuditLogConfig",
		"SentryHub",
		"LoggerFactory",
		"Database",
		"ConfigSourceController",
		"Resources",
	),
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
		"Request",
	),
	ProvideRequestContext,
	ProvideConfigSource,
	ProvideAppBaseResources,
	ProvideDatabaseConfig,
	ProvideAuditDatabaseCredentials,
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Value(template.DefaultLanguageTag(intl.DefaultLanguage)),
	wire.Value(template.SupportedLanguageTags([]string{intl.DefaultLanguage})),
)
