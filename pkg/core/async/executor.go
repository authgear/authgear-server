package async

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

type Executor struct {
	taskFactoryMap map[string]TaskFactory
	pool           db.Pool
}

func NewExecutor(dbPool db.Pool) *Executor {
	return &Executor{
		taskFactoryMap: map[string]TaskFactory{},
		pool:           dbPool,
	}
}

func (e *Executor) Register(name string, taskFactory TaskFactory) {
	e.taskFactoryMap[name] = taskFactory
}

func (e *Executor) Execute(taskCtx TaskContext, name string, param interface{}) {
	factory := e.taskFactoryMap[name]
	ctx := db.InitDBContext(context.Background(), e.pool)
	task := factory.NewTask(ctx, taskCtx)

	logHook := logging.NewDefaultLogHook(taskCtx.TenantConfig.DefaultSensitiveLoggerValues())
	sentryHook := &sentry.LogHook{Hub: sentry.DefaultClient.Hub}
	loggerFactory := logging.NewFactoryFromRequestID(taskCtx.RequestID, logHook, sentryHook)
	logger := loggerFactory.NewLogger("async-executor")
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.WithFields(map[string]interface{}{
					"task_name": name,
					"error":     rec,
				}).Error("unexpected error occurred when running async task")
			}
		}()

		err := task.Run(param)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"task_name": name,
				"error":     err,
			}).Error("error occurred when running async task")
		}
	}()
}
