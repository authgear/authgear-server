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
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var envConfigDeps = wire.NewSet(
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
		"StaticAssetURLPrefix",
		"Database",
	),
)

var rootDeps = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
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

var RequestDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
		"ResponseWriter",
	),
	ProvideRequestContext,
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
	),

	envConfigDeps,

	clock.DependencySet,
	globaldb.DependencySet,
	configsource.DependencySet,
)
