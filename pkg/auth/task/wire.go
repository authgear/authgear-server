package task

import (
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/task"
)

func newPwHouseKeeperTask(r *deps.TaskProvider) task.Task {
	return (*PwHousekeeperTask)(nil)
}

func newSendMessagesTask(r *deps.TaskProvider) task.Task {
	return (*SendMessagesTask)(nil)
}
