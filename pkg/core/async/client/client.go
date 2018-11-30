package client

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/async"

	"github.com/skygeario/skygear-server/pkg/core/async/server"
)

type TaskClient struct {
	context     context.Context
	taskContext async.TaskContext

	taskServer *server.TaskServer
}

func NewTaskClient(r *http.Request, taskServer *server.TaskServer) *TaskClient {
	return &TaskClient{
		context:     r.Context(),
		taskContext: async.NewTaskContext(r),
		taskServer:  taskServer,
	}
}

func (s *TaskClient) Submit(name string, param interface{}, response chan error) {
	if response == nil {
		s.taskServer.Handle(s.taskContext, name, param, nil)
		return
	}

	taskResponse := make(chan error)
	s.taskServer.Handle(s.taskContext, name, param, taskResponse)

	go func() {
		select {
		case <-s.context.Done(): // return if no one receive the error
		default:
			err := <-taskResponse
			response <- err
		}
	}()
}
