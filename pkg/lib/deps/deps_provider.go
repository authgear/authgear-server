package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var rootDeps = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
	),
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
		"StaticAssetURLPrefix",
	),

	ProvideCaptureTaskContext,
	ProvideRestoreTaskContext,

	clock.DependencySet,
	configsource.DependencySet,
)

var appRootDeps = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Config",
		"LoggerFactory",
		"Database",
		"Redis",
		"TaskQueue",
		"Resources",
	),

	wire.Bind(new(hook.DatabaseHandle), new(*db.Handle)),
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
