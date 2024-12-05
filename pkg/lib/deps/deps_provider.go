package deps

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var EnvConfigDeps = wire.NewSet(
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
		"AuthUISentryDSN",
		"AuthUIWindowMessageAllowedOrigins",
		"GlobalDatabase",
		"DatabaseConfig",
		"ImagesCDNHost",
		"WebAppCDNHost",
		"CORSAllowedOrigins",
		"AllowedFrameAncestors",
		"RedisConfig",
		"DenoEndpoint",
		"RateLimits",
		"SAML",
		"AppHostSuffixes",
		"UIImplementation",
		"UISettingsImplementation",
		"UserExportObjectStore",
		"SMSGatewayConfig",
	),
	wire.FieldsOf(new(*config.SMSGatewayEnvironmentConfig),
		"Default",
		"Twilio",
		"Custom",
		"Nexmo",
	),
	wire.FieldsOf(new(*config.SMSGatewayEnvironmentDefaultConfig),
		"UseConfigFrom",
		"Provider",
	),
)

var rootDeps = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"DatabasePool",
		"EmbeddedResources",
	),

	EnvConfigDeps,

	clock.DependencySet,
	globaldb.DependencySet,
	configsource.DependencySet,
)

var AppRootDeps = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"LoggerFactory",
		"AppDatabase",
		"SearchDatabase",
		"AuditReadDatabase",
		"AuditWriteDatabase",
		"Redis",
		"GlobalRedis",
		"AnalyticRedis",
		"AppContext",
	),
	wire.FieldsOf(new(*config.AppContext),
		"Resources",
		"Config",
		"Domains",
	),

	wire.Bind(new(event.Database), new(*appdb.Handle)),
	wire.Bind(new(workflow.ServiceDatabase), new(*appdb.Handle)),
	wire.Bind(new(authenticationflow.ServiceDatabase), new(*appdb.Handle)),
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(loginid.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.EmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),
	wire.Bind(new(hook.ResourceManager), new(*resource.Manager)),
)

var RootDependencySet = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*RootProvider),
		"LoggerFactory",
		"SentryHub",
		"BaseResources",
	),
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

func ProvideRedisQueueHTTPRequest() *http.Request {
	ctx := contextForRedisQueue
	r, _ := http.NewRequestWithContext(ctx, "GET", "", nil)
	return r
}

func ProvideRedisQueueRemoteIP() httputil.RemoteIP {
	return httputil.RemoteIP("127.0.0.1")
}

func ProvideRedisQueueUserAgentString() httputil.UserAgentString {
	return httputil.UserAgentString("redis-queue")
}

func ProvideRedisQueueHTTPHost() httputil.HTTPHost {
	return httputil.HTTPHost("127.0.0.1")
}

func ProvideRedisQueueHTTPProto() httputil.HTTPProto {
	return httputil.HTTPProto("https")
}

var RequestDependencySet = wire.NewSet(
	AppRootDeps,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
		"ResponseWriter",
	),
	ProvideRemoteIP,
	ProvideUserAgentString,
	ProvideHTTPHost,
	ProvideHTTPProto,
)

var RedisQueueDependencySet = wire.NewSet(
	AppRootDeps,
	ProvideRedisQueueHTTPRequest,
	ProvideRedisQueueRemoteIP,
	ProvideRedisQueueUserAgentString,
	ProvideRedisQueueHTTPHost,
	ProvideRedisQueueHTTPProto,
)

var BackgroundDependencySet = wire.NewSet(
	wire.FieldsOf(new(*BackgroundProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"LoggerFactory",
		"SentryHub",
		"DatabasePool",
		"RedisPool",
		"RedisHub",
		"BaseResources",
		"EmbeddedResources",
	),

	EnvConfigDeps,

	clock.DependencySet,
	globaldb.DependencySet,
)
