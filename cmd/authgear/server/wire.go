//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
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
