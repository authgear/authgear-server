package task

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/async/server"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify/verifycode"
)

const (
	// VerifyCodeSendTaskName provides the name for submiting VerifyCodeSendTask
	VerifyCodeSendTaskName = "VerifyCodeSendTask"
)

func AttachVerifyCodeSendTask(
	server *server.TaskServer,
	authDependency auth.DependencyMap,
) *server.TaskServer {
	server.Register(VerifyCodeSendTaskName, &VerifyCodeSendTaskFactory{
		authDependency,
	})
	return server
}

type VerifyCodeSendTaskFactory struct {
	DependencyMap auth.DependencyMap
}

func (c *VerifyCodeSendTaskFactory) NewTask(context async.TaskContext) async.Task {
	task := &VerifyCodeSendTask{}
	inject.DefaultTaskInject(task, c.DependencyMap, context)
	return async.TxTaskToTask(task, task.TxContext)
}

type VerifyCodeSendTask struct {
	CodeSenderFactory userverify.CodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	VerifyCodeStore   verifycode.Store             `dependency:"VerifyCodeStore"`
	TxContext         db.TxContext                 `dependency:"TxContext"`
	Logger            *logrus.Entry                `dependency:"HandlerLogger"`
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

	code := codeSender.Generate()
	if err = codeSender.Send(code, taskParam.Key, taskParam.Value, taskParam.UserProfile); err != nil {
		v.Logger.WithFields(logrus.Fields{
			"error":        err,
			"record_key":   taskParam.Key,
			"record_value": taskParam.Value,
		}).Error("fail to send verify request")
		return
	}

	verifyCode := verifycode.NewVerifyCode()
	verifyCode.UserID = taskParam.UserProfile.RecordID
	verifyCode.RecordKey = taskParam.Key
	verifyCode.RecordValue = taskParam.Value
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now()

	if err = v.VerifyCodeStore.CreateVerifyCode(&verifyCode); err != nil {
		return
	}

	return
}
