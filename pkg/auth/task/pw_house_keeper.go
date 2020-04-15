package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachPwHousekeeperTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) {
	// TODO(wire): fix task
	executor.Register(spec.PwHousekeeperTaskName, &PwHousekeeperTask{})
}

type PwHousekeeperTask struct {
	TxContext     db.TxContext         `dependency:"TxContext"`
	Logger        *logrus.Entry        `dependency:"HandlerLogger"`
	PwHousekeeper *audit.PwHousekeeper `dependency:"PwHousekeeper"`
}

func (t *PwHousekeeperTask) Run(ctx context.Context, param interface{}) (err error) {
	return db.WithTx(t.TxContext, func() error { return t.run(param) })
}

func (t *PwHousekeeperTask) run(param interface{}) (err error) {
	taskParam := param.(spec.PwHousekeeperTaskParam)

	t.Logger.WithFields(logrus.Fields{"user_id": taskParam.AuthID}).Debug("Housekeeping password")

	if err = taskParam.Validate(); err != nil {
		return
	}

	if err = t.PwHousekeeper.Housekeep(taskParam.AuthID); err != nil {
		return
	}
	return
}
