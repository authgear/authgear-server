package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/task/executors"
	taskqueue "github.com/skygeario/skygear-server/pkg/task/queue"
)

var appRootDeps = wire.NewSet(
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Context",
		"Config",
		"LoggerFactory",
		"DbContext",
		"RedisContext",
		"TemplateEngine",
	),
	wire.FieldsOf(new(*RootProvider),
		"ServerConfig",
		"TaskExecutor",
		"ReservedNameChecker",
	),

	ProvideCaptureTaskContext,
	wire.Bind(new(taskqueue.Executor), new(*executors.InMemoryExecutor)),
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
