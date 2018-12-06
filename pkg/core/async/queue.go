package async

import (
	"context"
	"net/http"
)

type Queue struct {
	context     context.Context
	taskContext TaskContext

	taskExecutor *Executor
}

func NewQueue(r *http.Request, taskExecutor *Executor) *Queue {
	return &Queue{
		context:      r.Context(),
		taskContext:  NewTaskContext(r),
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
