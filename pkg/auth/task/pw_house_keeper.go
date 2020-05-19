package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func AttachPwHousekeeperTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) {
	executor.Register(spec.PwHousekeeperTaskName, MakeTask(authDependency, newPwHouseKeeperTask))
}

type PwHousekeeperTask struct {
	TxContext     db.TxContext
	LoggerFactory logging.Factory
	PwHousekeeper *password.Housekeeper
}

func (t *PwHousekeeperTask) Run(ctx context.Context, param interface{}) (err error) {
	return db.WithTx(t.TxContext, func() error { return t.run(param) })
}

func (t *PwHousekeeperTask) run(param interface{}) (err error) {
	logger := t.LoggerFactory.NewLogger("passwordhousekeeper")
	taskParam := param.(spec.PwHousekeeperTaskParam)

	logger.WithFields(logrus.Fields{"user_id": taskParam.AuthID}).Debug("Housekeeping password")

	if err = taskParam.Validate(); err != nil {
		return
	}

	if err = t.PwHousekeeper.Housekeep(taskParam.AuthID); err != nil {
		return
	}
	return
}
