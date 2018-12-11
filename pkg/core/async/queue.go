package async

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Queue struct {
	context     context.Context
	taskContext TaskContext

	taskExecutor *Executor
}

func NewQueue(
	ctx context.Context,
	requestID string,
	tenantConfig config.TenantConfiguration,
	taskExecutor *Executor,
) *Queue {
	return &Queue{
		context: ctx,
		taskContext: TaskContext{
			RequestID:    requestID,
			TenantConfig: tenantConfig,
		},
		taskExecutor: taskExecutor,
	}
}

func (s *Queue) Enqueue(name string, param interface{}, response chan error) {
	if response == nil {
		s.taskExecutor.Execute(s.taskContext, name, param, nil)
		return
	}

	taskResponse := make(chan error)
	s.taskExecutor.Execute(s.taskContext, name, param, taskResponse)

	go func() {
		select {
		case <-s.context.Done(): // return if no one receive the error
		default:
			err := <-taskResponse
			response <- err
		}
	}()
}
