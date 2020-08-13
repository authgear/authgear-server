package tasks

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigurePwHousekeeperTask(registry task.Registry, t task.Task) {
	registry.Register(tasks.PwHousekeeper, t)
}

type PwHousekeeperLogger struct{ *log.Logger }

func NewPwHousekeeperLogger(lf *log.Factory) PwHousekeeperLogger {
	return PwHousekeeperLogger{lf.New("password-housekeeper")}
}

type PwHousekeeperTask struct {
	Database      *db.Handle
	Logger        PwHousekeeperLogger
	PwHousekeeper *password.Housekeeper
}

func (t *PwHousekeeperTask) Run(ctx context.Context, param task.Param) (err error) {
	return t.Database.WithTx(func() error { return t.run(param) })
}

func (t *PwHousekeeperTask) run(param task.Param) (err error) {
	taskParam := param.(*tasks.PwHousekeeperParam)

	t.Logger.WithFields(logrus.Fields{"user_id": taskParam.UserID}).Debug("Housekeeping password")

	if err = taskParam.Validate(); err != nil {
		return
	}

	if err = t.PwHousekeeper.Housekeep(taskParam.UserID); err != nil {
		return
	}
	return
}
