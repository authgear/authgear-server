package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"

	configsource "github.com/authgear/authgear-server/pkg/lib/config/source"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

var rootDeps = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"ServerConfig",
		"ReservedNameChecker",
	),

	ProvideCaptureTaskContext,
	ProvideRestoreTaskContext,

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
		"TemplateEngine",
	),

	wire.Bind(new(hook.DatabaseHandle), new(*db.Handle)),
)

var RootDependencySet = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*RootProvider),
		"LoggerFactory",
		"SentryHub",
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
	requestDeps,
)

var TaskDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*TaskProvider),
		"AppProvider",
		"Context",
	),
	taskDeps,
)
