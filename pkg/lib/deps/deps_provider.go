package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var envConfigDeps = wire.NewSet(
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
		"GlobalDatabase",
		"DatabaseConfig",
		"ImagesCDNHost",
		"WebAppCDNHost",
		"CORSAllowedOrigins",
		"RedisConfig",
		"NFTIndexerAPIEndpoint",
	),
)

var rootDeps = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"DatabasePool",
		"EmbeddedResources",
	),

	envConfigDeps,

	ProvideCaptureTaskContext,
	ProvideRestoreTaskContext,

	clock.DependencySet,
	globaldb.DependencySet,
	configsource.DependencySet,
)

var appRootDeps = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Config",
		"LoggerFactory",
		"AppDatabase",
		"AuditReadDatabase",
		"AuditWriteDatabase",
		"Redis",
		"AnalyticRedis",
		"TaskQueue",
		"Resources",
	),

	wire.Bind(new(event.Database), new(*appdb.Handle)),
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(loginid.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.EmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),
)

var RootDependencySet = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*RootProvider),
		"LoggerFactory",
		"SentryHub",
		"BaseResources",
	),
)

func ProvideRequestContext(r *http.Request) context.Context { return r.Context() }

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

var RequestDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
		"ResponseWriter",
	),
	ProvideRequestContext,
	ProvideRemoteIP,
	ProvideUserAgentString,
	ProvideHTTPHost,
	ProvideHTTPProto,
)

var TaskDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*TaskProvider),
		"AppProvider",
		"Context",
	),
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

	envConfigDeps,

	clock.DependencySet,
	globaldb.DependencySet,
)
