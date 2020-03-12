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
	"github.com/skygeario/skygear-server/pkg/core/inject"
)

func AttachWelcomeEmailSendTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) *async.Executor {
	executor.Register(spec.WelcomeEmailSendTaskName, &WelcomeEmailSendTaskFactory{
		authDependency,
	})
	return executor
}

type WelcomeEmailSendTaskFactory struct {
	DependencyMap auth.DependencyMap
}

func (c *WelcomeEmailSendTaskFactory) NewTask(ctx context.Context, taskCtx async.TaskContext) async.Task {
	task := &WelcomeEmailSendTask{}
	inject.DefaultTaskInject(task, c.DependencyMap, ctx, taskCtx)
	return async.TxTaskToTask(task, task.TxContext)
}

type WelcomeEmailSendTask struct {
	WelcomeEmailSender welcemail.Sender  `dependency:"WelcomeEmailSender"`
	UserProfileStore   userprofile.Store `dependency:"UserProfileStore"`
	TxContext          db.TxContext      `dependency:"TxContext"`
	Logger             *logrus.Entry     `dependency:"HandlerLogger"`
}

func (w *WelcomeEmailSendTask) WithTx() bool {
	return true
}

func (w *WelcomeEmailSendTask) Run(param interface{}) (err error) {
	taskParam := param.(spec.WelcomeEmailSendTaskParam)

	w.Logger.WithFields(logrus.Fields{"user_id": taskParam.User.ID}).Debug("Sending welcome email")

	if err = w.WelcomeEmailSender.Send(taskParam.URLPrefix, taskParam.Email, taskParam.User); err != nil {
		err = errors.WithDetails(err, errors.Details{"user_id": taskParam.User.ID})
		return
	}

	return
}
