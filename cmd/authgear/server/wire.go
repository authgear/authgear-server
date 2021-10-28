//go:build wireinject
// +build wireinject

package server

import (
	"context"
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
)

func newConfigSourceController(p *deps.RootProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		deps.RootDependencySet,
		wire.FieldsOf(new(*deps.RootProvider),
			"DatabasePool",
		),
	))
}

func newInProcessQueue(p *deps.AppProvider, e *executor.InProcessExecutor) *queue.InProcessQueue {
	panic(wire.Build(
		deps.RootDependencySet,
		wire.FieldsOf(new(*deps.AppProvider),
			"Config",
			"AppDatabase",
		),
		queue.DependencySet,
		wire.Bind(new(queue.Executor), new(*executor.InProcessExecutor)),
	))
}
