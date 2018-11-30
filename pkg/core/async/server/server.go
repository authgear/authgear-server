package server

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

type TaskServer struct {
	taskFactoryMap map[string]async.TaskFactory
}

func NewTaskServer() *TaskServer {
	return &TaskServer{
		taskFactoryMap: map[string]async.TaskFactory{},
	}
}

func (t *TaskServer) Register(name string, taskFactory async.TaskFactory) {
	t.taskFactoryMap[name] = taskFactory
}

func (t *TaskServer) Handle(ctx async.TaskContext, name string, param interface{}, response chan error) {
	factory := t.taskFactoryMap[name]
	task := factory.NewTask(ctx)

	formatter := logging.CreateMaskFormatter(ctx.TenantConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
	logger := logging.CreateLoggerWithRequestID(ctx.RequestID, "async_task_server", formatter)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.WithFields(map[string]interface{}{
					"task_name": name,
					"error":     rec,
				}).Error("unexpected error occurred when running async task")

				if response != nil {
					switch err := rec.(type) {
					case error:
						response <- err
					default:
						response <- fmt.Errorf("%+v", err)
					}
				}
			}
		}()

		err := task.Run(param)
		if response != nil {
			response <- err
		}
	}()
}
