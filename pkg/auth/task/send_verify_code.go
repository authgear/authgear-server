package task

import (
	"context"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
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
	CodeSenderFactory    userverify.CodeSenderFactory    `dependency:"UserVerifyCodeSenderFactory"`
	CodeGeneratorFactory userverify.CodeGeneratorFactory `dependency:"VerifyCodeCodeGeneratorFactory"`
	VerifyCodeStore      userverify.Store                `dependency:"VerifyCodeStore"`
	TxContext            db.TxContext                    `dependency:"TxContext"`
	Logger               *logrus.Entry                   `dependency:"HandlerLogger"`
}

type VerifyCodeSendTaskParam struct {
	Key         string
	Value       string
	UserProfile userprofile.UserProfile
}

func (v *VerifyCodeSendTask) WithTx() bool {
	return true
}

func (v *VerifyCodeSendTask) Run(param interface{}) (err error) {
	taskParam := param.(VerifyCodeSendTaskParam)
	codeSender := v.CodeSenderFactory.NewCodeSender(taskParam.Key)

	v.Logger.WithFields(logrus.Fields{
		"userID": taskParam.UserProfile.ID,
	}).Info("start sending user verify requests")

	codeGenerator := v.CodeGeneratorFactory.NewCodeGenerator(taskParam.Key)
	code := codeGenerator.Generate()

	verifyCode := userverify.NewVerifyCode()
	verifyCode.UserID = taskParam.UserProfile.ID
	verifyCode.RecordKey = taskParam.Key
	verifyCode.RecordValue = taskParam.Value
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now()

	if err = v.VerifyCodeStore.CreateVerifyCode(&verifyCode); err != nil {
		return
	}

	if err = codeSender.Send(verifyCode, taskParam.UserProfile); err != nil {
		v.Logger.WithFields(logrus.Fields{
			"error":        err,
			"record_key":   taskParam.Key,
			"record_value": taskParam.Value,
		}).Error("fail to send verify request")
		return
	}

	return
}
