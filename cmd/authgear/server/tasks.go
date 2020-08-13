package server

import (
	authtask "github.com/authgear/authgear-server/pkg/auth/task"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

func setupTasks(registry task.Registry, p *deps.RootProvider) {
	authtask.ConfigurePwHousekeeperTask(registry, p.Task(newPwHousekeeperTask))
	authtask.ConfigureSendMessagesTask(registry, p.Task(newSendMessagesTask))
}
