package deps

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ProvideRemoteIP(r *http.Request, trustProxy config.TrustProxy) httputil.RemoteIP {
	return httputil.RemoteIP(httputil.GetIP(r, bool(trustProxy)))
}

func ProvideHTTPHost(r *http.Request, trustProxy config.TrustProxy) httputil.HTTPHost {
	return httputil.HTTPHost(httputil.GetHost(r, bool(trustProxy)))
}

func ProvideHTTPProto(r *http.Request, trustProxy config.TrustProxy) httputil.HTTPProto {
	return httputil.HTTPProto(httputil.GetProto(r, bool(trustProxy)))
}

func ProvideUserAgentString(r *http.Request) httputil.UserAgentString {
	return httputil.UserAgentString(r.UserAgent())
}

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

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"AuthgearConfig",
		"AdminAPIConfig",
		"AppConfig",
		"SMTPConfig",
		"MailConfig",
		"KubernetesConfig",
		"DomainImplementation",
		"SearchConfig",
		"AuditLogConfig",
		"AnalyticConfig",
		"StripeConfig",
		"OsanoConfig",
		"GoogleTagManagerConfig",
		"PortalFrontendSentryConfig",
		"PortalFeaturesConfig",
		"SentryHub",
		"Database",
		"RedisPool",
		"ConfigSourceController",
		"Resources",
		"FilesystemCache",
	),
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
		"GlobalDatabase",
		"GlobalRedis",
		"DatabaseConfig",
		"RedisConfig",
		"DenoEndpoint",
		"AppHostSuffixes",
		"UIImplementation",
		"UISettingsImplementation",
		"SAML",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
		"Request",
	),
	ProvideRemoteIP,
	ProvideUserAgentString,
	ProvideHTTPHost,
	ProvideHTTPProto,
	ProvideConfigSource,
	ProvideAppBaseResources,
	ProvideAuditDatabaseCredentials,
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Value(template.DefaultLanguageTag(intl.BuiltinBaseLanguage)),
	wire.Value(template.SupportedLanguageTags([]string{intl.BuiltinBaseLanguage})),
)
