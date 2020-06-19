package task

import (
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/task"
)

func newPwHouseKeeperTask(r *deps.RequestProvider) task.Task {
	return (*PwHousekeeperTask)(nil)
}

func newSendMessagesTask(r *deps.RequestProvider) task.Task {
	return (*SendMessagesTask)(nil)
}
