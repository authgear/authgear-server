package task

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/portal/task/tasks"
)

var DependencySet = wire.NewSet(
	NewInProcessExecutorLogger,
	NewExecutor,

	wire.Bind(new(Executor), new(*InProcessExecutor)),
	wire.Struct(new(InProcessQueue), "*"),
)

func NewExecutor(
	logger InProcessExecutorLogger,
	sendMessageTask *tasks.SendMessagesTask,
) *InProcessExecutor {
	executor := &InProcessExecutor{
		Logger: logger,
	}
	tasks.ConfigureSendMessagesTask(executor, sendMessageTask)
	return executor
}
