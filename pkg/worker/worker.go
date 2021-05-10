package worker

import (
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/worker/tasks"
)

type Worker struct {
	Executor *executor.InProcessExecutor
}

func NewWorker(provider *deps.RootProvider) *Worker {
	executor := newInProcessExecutor(provider)
	tasks.ConfigureSendMessagesTask(executor, provider.Task(newSendMessagesTask))
	tasks.ConfigureReindexUserTask(executor, provider.Task(newReindexUserTask))
	return &Worker{Executor: executor}
}
