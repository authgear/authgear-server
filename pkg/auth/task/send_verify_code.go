package task

import (
	"context"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/response"
)

const (
	// VerifyCodeSendTaskName provides the name for submiting VerifyCodeSendTask
	VerifyCodeSendTaskName = "VerifyCodeSendTask"
)

func AttachVerifyCodeSendTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) *async.Executor {
	executor.Register(VerifyCodeSendTaskName, &VerifyCodeSendTaskFactory{
		authDependency,
	})
	return executor
}

type VerifyCodeSendTaskFactory struct {
	DependencyMap auth.DependencyMap
}

func (c *VerifyCodeSendTaskFactory) NewTask(ctx context.Context, taskCtx async.TaskContext) async.Task {
	task := &VerifyCodeSendTask{}
	inject.DefaultTaskInject(task, c.DependencyMap, ctx, taskCtx)
	return async.TxTaskToTask(task, task.TxContext)
}

type VerifyCodeSendTask struct {
	CodeSenderFactory userverify.CodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	CodeGenerator     userverify.CodeGenerator     `dependency:"VerifyCodeCodeGenerator"`
	VerifyCodeStore   userverify.Store             `dependency:"VerifyCodeStore"`
	TxContext         db.TxContext                 `dependency:"TxContext"`
	Logger            *logrus.Entry                `dependency:"HandlerLogger"`
}

type VerifyCodeSendTaskParam struct {
	LoginIDKey string
	LoginID    string
	User       response.User
}

func (v *VerifyCodeSendTask) WithTx() bool {
	return true
}

func (v *VerifyCodeSendTask) Run(param interface{}) (err error) {
	taskParam := param.(VerifyCodeSendTaskParam)
	codeSender := v.CodeSenderFactory.NewCodeSender(taskParam.LoginIDKey)

	v.Logger.WithFields(logrus.Fields{
		"user_id": taskParam.User.UserID,
	}).Info("start sending user verify requests")

	code := v.CodeGenerator.Generate(taskParam.LoginIDKey)

	verifyCode := userverify.NewVerifyCode()
	verifyCode.UserID = taskParam.User.UserID
	verifyCode.LoginIDKey = taskParam.LoginIDKey
	verifyCode.LoginID = taskParam.LoginID
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now()

	if err = v.VerifyCodeStore.CreateVerifyCode(&verifyCode); err != nil {
		return
	}

	if err = codeSender.Send(verifyCode, taskParam.User); err != nil {
		v.Logger.WithFields(logrus.Fields{
			"error":        err,
			"login_id_key": taskParam.LoginIDKey,
			"login_id":     taskParam.LoginID,
		}).Error("fail to send verify request")
		return
	}

	return
}
