//+build wireinject

package worker

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	workertasks "github.com/authgear/authgear-server/pkg/worker/tasks"
)

func newInProcessExecutor(p *deps.RootProvider) *executor.InProcessExecutor {
	panic(wire.Build(
		deps.RootDependencySet,
		executor.DependencySet,
	))
}

func newSendMessagesTask(p *deps.TaskProvider) task.Task {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(task.Task), new(*workertasks.SendMessagesTask)),
	))
}

func newReindexUserTask(p *deps.TaskProvider) task.Task {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(task.Task), new(*workertasks.ReindexUserTask)),
	))
}
