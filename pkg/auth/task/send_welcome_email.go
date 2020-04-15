package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

func AttachWelcomeEmailSendTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) {
	// TODO(wire): fix task
	executor.Register(spec.WelcomeEmailSendTaskName, &WelcomeEmailSendTask{})
}

type WelcomeEmailSendTask struct {
	WelcomeEmailSender welcemail.Sender  `dependency:"WelcomeEmailSender"`
	UserProfileStore   userprofile.Store `dependency:"UserProfileStore"`
	TxContext          db.TxContext      `dependency:"TxContext"`
	Logger             *logrus.Entry     `dependency:"HandlerLogger"`
}

func (w *WelcomeEmailSendTask) Run(ctx context.Context, param interface{}) (err error) {
	return db.WithTx(w.TxContext, func() error { return w.run(param) })
}

func (w *WelcomeEmailSendTask) run(param interface{}) (err error) {
	taskParam := param.(spec.WelcomeEmailSendTaskParam)

	w.Logger.WithFields(logrus.Fields{"user_id": taskParam.User.ID}).Debug("Sending welcome email")

	if err = w.WelcomeEmailSender.Send(taskParam.URLPrefix, taskParam.Email, taskParam.User); err != nil {
		err = errors.WithDetails(err, errors.Details{"user_id": taskParam.User.ID})
		return
	}

	return
}
