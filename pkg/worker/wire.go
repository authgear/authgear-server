//+build wireinject

package worker

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	authtask "github.com/authgear/authgear-server/pkg/worker/tasks"
)

func newInProcessExecutor(p *deps.RootProvider) *executor.InProcessExecutor {
	panic(wire.Build(
		deps.RootDependencySet,
		executor.DependencySet,
	))
}

func newPwHousekeeperTask(p *deps.TaskProvider) task.Task {
	panic(wire.Build(
		deps.TaskDependencySet,
		wire.Bind(new(task.Task), new(*authtask.PwHousekeeperTask)),
	))
}

func newSendMessagesTask(p *deps.TaskProvider) task.Task {
	panic(wire.Build(
		deps.TaskDependencySet,
		wire.Bind(new(task.Task), new(*authtask.SendMessagesTask)),
	))
}
