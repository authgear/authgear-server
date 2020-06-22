package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/task"
)

func AttachPwHousekeeperTask(
	registry task.Registry,
	p *deps.RootProvider,
) {
	registry.Register(spec.PwHousekeeperTaskName, p.Task(newPwHouseKeeperTask))
}

type PwHousekeeperTask struct {
	DBContext     db.Context
	LoggerFactory logging.Factory
	PwHousekeeper *password.Housekeeper
}

func (t *PwHousekeeperTask) Run(ctx context.Context, param interface{}) (err error) {
	return db.WithTx(t.DBContext, func() error { return t.run(param) })
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
