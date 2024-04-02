//go:build wireinject
// +build wireinject

package e2e

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

func newConfigSourceController(p *deps.RootProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		configsource.NewResolveAppIDTypeDomain,
		deps.RootDependencySet,
		configsource.ControllerDependencySet,
	))
}

func newInProcessQueue(p *deps.AppProvider, e *executor.InProcessExecutor) *queue.InProcessQueue {
	panic(wire.Build(
		deps.RootDependencySet,
		wire.FieldsOf(new(*deps.AppProvider),
			"AppContext",
			"AppDatabase",
		),
		wire.FieldsOf(new(*config.AppContext),
			"Config",
		),
		queue.DependencySet,
		wire.Bind(new(queue.Executor), new(*executor.InProcessExecutor)),
	))
}

func newUserImport(p *deps.AppProvider, c context.Context) *userimport.UserImportService {
	panic(wire.Build(
		deps.End2EndDependencySet,
		deps.CommonDependencySet,
	))
}
