package task

import (
	"context"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/model"
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
	CodeSenderFactory        userverify.CodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	AuthInfoStore            authinfo.Store               `dependency:"AuthInfoStore"`
	UserProfileStore         userprofile.Store            `dependency:"UserProfileStore"`
	UserVerificationProvider userverify.Provider          `dependency:"UserVerificationProvider"`
	PasswordAuthProvider     password.Provider            `dependency:"PasswordAuthProvider"`
	IdentityProvider         principal.IdentityProvider   `dependency:"IdentityProvider"`
	TxContext                db.TxContext                 `dependency:"TxContext"`
	Logger                   *logrus.Entry                `dependency:"HandlerLogger"`
}

type VerifyCodeSendTaskParam struct {
	LoginID string
	UserID  string
}

func (v *VerifyCodeSendTask) WithTx() bool {
	return true
}

func (v *VerifyCodeSendTask) Run(param interface{}) (err error) {
	taskParam := param.(VerifyCodeSendTaskParam)
	loginID := taskParam.LoginID
	userID := taskParam.UserID

	v.Logger.WithFields(logrus.Fields{
		"login_id": loginID,
		"user_id":  userID,
	}).Info("start sending user verify message")

	authInfo := authinfo.AuthInfo{}
	err = v.AuthInfoStore.GetAuth(userID, &authInfo)
	if err != nil {
		err = skyerr.NewError(skyerr.UnexpectedError, "unable to fetch user")
		return
	}

	userProfile, err := v.UserProfileStore.GetUserProfile(userID)
	if err != nil {
		err = skyerr.NewError(skyerr.UnexpectedError, "unable to fetch user profile")
		return
	}

	// We don't check realms. i.e. Verifying a email means every email login IDs
	// of that email is verified, regardless the realm.
	principals, err := v.PasswordAuthProvider.GetPrincipalsByLoginID("", loginID)
	if err != nil {
		return
	}

	var userPrincipal *password.Principal
	for _, principal := range principals {
		if principal.UserID == authInfo.ID {
			userPrincipal = principal
			break
		}
	}
	if userPrincipal == nil {
		err = skyerr.NewError(skyerr.UnexpectedError, "Value of "+loginID+" doesn't exist.")
		return
	}

	verifyCode, err := v.UserVerificationProvider.CreateVerifyCode(userPrincipal)
	if err != nil {
		return
	}

	codeSender := v.CodeSenderFactory.NewCodeSender(userPrincipal.LoginIDKey)
	user := model.NewUser(authInfo, userProfile)
	if err = codeSender.Send(*verifyCode, user); err != nil {
		v.Logger.WithFields(logrus.Fields{
			"error":        err,
			"login_id_key": userPrincipal.LoginIDKey,
			"login_id":     userPrincipal.LoginID,
		}).Error("fail to send verify request")
		return
	}

	return nil
}
