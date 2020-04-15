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
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func AttachWelcomeEmailSendTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) {
	executor.Register(spec.WelcomeEmailSendTaskName, MakeTask(authDependency, newWelcomeEmailSendTask))
}

type WelcomeEmailSendTask struct {
	WelcomeEmailSender welcemail.Sender
	UserProfileStore   userprofile.Store
	TxContext          db.TxContext
	LoggerFactory      logging.Factory
}

func (w *WelcomeEmailSendTask) Run(ctx context.Context, param interface{}) (err error) {
	return db.WithTx(w.TxContext, func() error { return w.run(param) })
}

func (w *WelcomeEmailSendTask) run(param interface{}) (err error) {
	taskParam := param.(spec.WelcomeEmailSendTaskParam)

	logger := w.LoggerFactory.NewLogger("welcomeemail")

	logger.WithFields(logrus.Fields{"user_id": taskParam.User.ID}).Debug("Sending welcome email")

	if err = w.WelcomeEmailSender.Send(taskParam.URLPrefix, taskParam.Email, taskParam.User); err != nil {
		err = errors.WithDetails(err, errors.Details{"user_id": taskParam.User.ID})
		return
	}

	return
}
