package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func AttachVerifyCodeSendTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) {
	executor.Register(spec.VerifyCodeSendTaskName, MakeTask(authDependency, newVerifyCodeSendTask))
}

type VerifyCodeLoginIDProvider interface {
	GetByLoginID(loginid.LoginID) ([]*loginid.Identity, error)
}

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type VerifyCodeSendTask struct {
	CodeSenderFactory        userverify.CodeSenderFactory
	Users                    UserProvider
	UserVerificationProvider userverify.Provider
	LoginIDProvider          VerifyCodeLoginIDProvider
	TxContext                db.TxContext
	LoggerFactory            logging.Factory
}

func (v *VerifyCodeSendTask) Run(ctx context.Context, param interface{}) (err error) {
	return db.WithTx(v.TxContext, func() error { return v.run(param) })
}

func (v *VerifyCodeSendTask) run(param interface{}) (err error) {
	taskParam := param.(spec.VerifyCodeSendTaskParam)
	loginID := taskParam.LoginID
	userID := taskParam.UserID

	logger := v.LoggerFactory.NewLogger("verifycode")
	logger.WithFields(logrus.Fields{"user_id": taskParam.UserID}).Debug("Sending verification code")

	user, err := v.Users.Get(userID)
	if err != nil {
		return
	}

	is, err := v.LoginIDProvider.GetByLoginID(loginid.LoginID{Value: loginID})
	if err != nil {
		return
	}

	var identity *loginid.Identity
	for _, i := range is {
		if i.UserID == user.ID {
			identity = i
			break
		}
	}
	if identity == nil {
		err = errors.WithDetails(errors.New("login ID not found"), errors.Details{"user_id": userID})
		return
	}

	verifyCode, err := v.UserVerificationProvider.CreateVerifyCode(identity)
	if err != nil {
		return
	}

	codeSender := v.CodeSenderFactory.NewCodeSender(taskParam.URLPrefix, identity.LoginIDKey)
	if err = codeSender.Send(*verifyCode, *user); err != nil {
		err = errors.WithDetails(err, errors.Details{"user_id": userID})
		return
	}

	return nil
}
