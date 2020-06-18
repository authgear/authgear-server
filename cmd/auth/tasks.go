package main

import (
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
)

// nolint: deadcode
func setupTasks(executor *async.Executor, deps auth.DependencyMap) {
	task.AttachPwHousekeeperTask(executor, deps)
	task.AttachSendMessagesTask(executor, deps)
}
