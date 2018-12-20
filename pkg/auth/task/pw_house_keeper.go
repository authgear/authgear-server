package task

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"
)

const (
	// PwHousekeeperTaskName provides the name for submiting PwHousekeeperTask
	PwHousekeeperTaskName = "PwHousekeeperTask"
)

func AttachPwHousekeeperTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) *async.Executor {
	executor.Register(PwHousekeeperTaskName, &PwHousekeeperTaskFactory{
		authDependency,
	})
	return executor
}

type PwHousekeeperTaskFactory struct {
	DependencyMap auth.DependencyMap
}

func (f *PwHousekeeperTaskFactory) NewTask(ctx context.Context, taskCtx async.TaskContext) async.Task {
	task := &PwHousekeeperTask{}
	inject.DefaultTaskInject(task, f.DependencyMap, ctx, taskCtx)
	return async.TxTaskToTask(task, task.TxContext)
}

type PwHousekeeperTask struct {
	TxContext     db.TxContext         `dependency:"TxContext"`
	Logger        *logrus.Entry        `dependency:"HandlerLogger"`
	PwHousekeeper *audit.PwHousekeeper `dependency:"PwHousekeeper"`
}

type PwHousekeeperTaskParam struct {
	AuthID string
}

func (p PwHousekeeperTaskParam) Validate() error {
	if p.AuthID == "" {
		return errors.New("missing auth id for pw house keeping task")
	}

	return nil
}

func (t *PwHousekeeperTask) WithTx() bool {
	return true
}

func (t *PwHousekeeperTask) Run(param interface{}) (err error) {
	taskParam := param.(PwHousekeeperTaskParam)

	if err = taskParam.Validate(); err != nil {
		return
	}

	if err = t.PwHousekeeper.Housekeep(taskParam.AuthID); err != nil {
		return
	}

	return
}
