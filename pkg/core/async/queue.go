package async

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Queue interface {
	Enqueue(name string, param interface{}, response chan error)
}

type queue struct {
	context     context.Context
	taskContext TaskContext

	taskExecutor *Executor
}

func NewQueue(
	ctx context.Context,
	requestID string,
	tenantConfig config.TenantConfiguration,
	taskExecutor *Executor,
) Queue {
	return &queue{
		context: ctx,
		taskContext: TaskContext{
			RequestID:    requestID,
			TenantConfig: tenantConfig,
		},
		taskExecutor: taskExecutor,
	}
}

func (s *queue) Enqueue(name string, param interface{}, response chan error) {
	if response == nil {
		s.taskExecutor.Execute(s.taskContext, name, param, nil)
		return
	}

	taskResponse := make(chan error)

	go func() {
		err := <-taskResponse
		select {
		case <-s.context.Done(): // return if no one receive the error
		case response <- err:
		}
	}()

	s.taskExecutor.Execute(s.taskContext, name, param, taskResponse)
}
