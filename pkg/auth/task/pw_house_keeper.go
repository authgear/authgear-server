package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/auth/task/spec"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigurePwHousekeeperTask(registry task.Registry, t task.Task) {
	registry.Register(spec.PwHousekeeperTaskName, t)
}

type PwHousekeeperLogger struct{ *log.Logger }

func NewPwHousekeeperLogger(lf *log.Factory) PwHousekeeperLogger {
	return PwHousekeeperLogger{lf.New("password_housekeeper")}
}

type PwHousekeeperTask struct {
	Database      *db.Handle
	Logger        PwHousekeeperLogger
	PwHousekeeper *password.Housekeeper
}

func (t *PwHousekeeperTask) Run(ctx context.Context, param interface{}) (err error) {
	return t.Database.WithTx(func() error { return t.run(param) })
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
