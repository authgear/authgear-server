package executor

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type InProcessExecutorLogger struct{ *log.Logger }

func NewInProcessExecutorLogger(lf *log.Factory) InProcessExecutorLogger {
	return InProcessExecutorLogger{lf.New("task-executor")}
}

type InProcessExecutor struct {
	Logger         InProcessExecutorLogger
	RestoreContext task.RestoreTaskContext

	tasks map[string]task.Task `wire:"-"`
}

func (e *InProcessExecutor) Register(name string, t task.Task) {
	if e.tasks == nil {
		e.tasks = map[string]task.Task{}
	}
	e.tasks[name] = t
}

func (e *InProcessExecutor) Run(taskCtx *task.Context, param task.Param) {
	ctx := e.RestoreContext(context.Background(), taskCtx)
	task := e.tasks[param.TaskName()]

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				e.Logger.WithFields(map[string]interface{}{
					"task_name": param.TaskName(),
					"error":     rec,
					"stack":     errorutil.Callers(8),
				}).Error("unexpected error occurred when running async task")
			}
		}()

		err := task.Run(ctx, param)
		if err != nil {
			e.Logger.WithFields(map[string]interface{}{
				"task_name": param.TaskName(),
				"error":     err,
				"stack":     errorutil.Callers(8),
			}).Error("error occurred when running async task")
		}
	}()
}
