package main

import (
	authtask "github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/task"
)

func setupTasks(registry task.Registry, p *deps.RootProvider) {
	authtask.ConfigurePwHousekeeperTask(registry, nil)
	authtask.ConfigureSendMessagesTask(registry, nil)
}
