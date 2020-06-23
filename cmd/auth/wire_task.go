//+build wireinject

package main

import (
	"github.com/google/wire"

	authtask "github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/task"
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
