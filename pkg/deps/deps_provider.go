package deps

import (
	"github.com/google/wire"

	configsource "github.com/authgear/authgear-server/pkg/auth/config/source"
	"github.com/authgear/authgear-server/pkg/task/executors"
	taskqueue "github.com/authgear/authgear-server/pkg/task/queue"
)

var rootDeps = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"ServerConfig",
		"TaskExecutor",
		"ReservedNameChecker",
	),

	ProvideCaptureTaskContext,
	wire.Bind(new(taskqueue.Executor), new(*executors.InMemoryExecutor)),

	configsource.DependencySet,
)

var appRootDeps = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Context",
		"Config",
		"LoggerFactory",
		"Database",
		"Redis",
		"TemplateEngine",
	),
)

var RootDependencySet = wire.NewSet(
	rootDeps,
	wire.FieldsOf(new(*RootProvider),
		"LoggerFactory",
	),
)

var RequestDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
	),
	requestDeps,
)

var TaskDependencySet = wire.NewSet(
	appRootDeps,
	wire.FieldsOf(new(*TaskProvider),
		"AppProvider",
	),
	taskDeps,
)
