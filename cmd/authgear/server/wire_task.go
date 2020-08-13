//+build wireinject

package server

import (
	"github.com/google/wire"

	authtask "github.com/authgear/authgear-server/pkg/auth/task"
	"github.com/authgear/authgear-server/pkg/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

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
