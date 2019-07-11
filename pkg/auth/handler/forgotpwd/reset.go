package forgotpwd

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachForgotPasswordResetHandler attaches ForgotPasswordResetHandler to server
func AttachForgotPasswordResetHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/forgot_password/reset_password", &ForgotPasswordResetHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	server.Handle("/forgot_password/reset_password_form", &ForgotPasswordResetFormHandlerFactory{
		authDependency,
	}).Methods("POST", "GET")
	return server
}

// ForgotPasswordResetHandlerFactory creates ForgotPasswordResetHandler
type ForgotPasswordResetHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordResetHandler
func (f ForgotPasswordResetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordResetHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ForgotPasswordResetHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type ForgotPasswordResetPayload struct {
	UserID       string `json:"user_id"`
	Code         string `json:"code"`
	ExpireAt     int64  `json:"expire_at"`
	ExpireAtTime time.Time
	NewPassword  string `json:"new_password"`
}

func (payload ForgotPasswordResetPayload) Validate() error {
	if payload.UserID == "" {
		return skyerr.NewInvalidArgument("empty user_id", []string{"user_id"})
	}

	if payload.Code == "" {
		return skyerr.NewInvalidArgument("empty code", []string{"code"})
	}

	if payload.ExpireAt == 0 {
		return skyerr.NewInvalidArgument("empty expire_at", []string{"expire_at"})
	}

	if payload.NewPassword == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"new_password"})
	}

	return nil
}

// ForgotPasswordResetHandler reset user password with given code from email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/forgot_password/reset_password <<EOF
//  {
//    "user_id": "xxx",
//    "code": "xxx",
//    "expire_at": xxx, (utc timestamp)
//    "new_password": "xxx",
//  }
//  EOF
type ForgotPasswordResetHandler struct {
	CodeGenerator        *forgotpwdemail.CodeGenerator `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker      *authAudit.PasswordChecker    `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store               `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store                `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider             `dependency:"PasswordAuthProvider"`
	AuditTrail           audit.Trail                   `dependency:"AuditTrail"`
	TxContext            db.TxContext                  `dependency:"TxContext"`
	Logger               *logrus.Entry                 `dependency:"HandlerLogger"`
	TaskQueue            async.Queue                   `dependency:"AsyncTaskQueue"`
}

func (h ForgotPasswordResetHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ForgotPasswordResetHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ForgotPasswordResetPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	payload.ExpireAtTime = time.Unix(payload.ExpireAt, 0).UTC()

	return payload, nil
}

func (h ForgotPasswordResetHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ForgotPasswordResetPayload)

	// check code expiration
	if timeNow().After(payload.ExpireAtTime) {
		h.Logger.Error("forgot password code expired")
		err = genericResetPasswordError()
		return
	}

	authInfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(payload.UserID, &authInfo); e != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(e).Error("user not found")
		err = genericResetPasswordError()
		return
	}

	// Get password auth principals
	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get password auth principals")
		err = genericResetPasswordError()
		return
	}

	// Get user email from loginIDs
	hashedPassword := principals[0].HashedPassword
	expectedCode := h.CodeGenerator.Generate(authInfo, hashedPassword, payload.ExpireAtTime)
	if payload.Code != expectedCode {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":       payload.UserID,
			"code":          payload.Code,
			"expected_code": expectedCode,
		}).Error("wrong forgot password reset password code")
		err = genericResetPasswordError()
		return
	}

	resetPwdCtx := password.ResetPasswordRequestContext{
		PasswordChecker:      h.PasswordChecker,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}

	if err = resetPwdCtx.ExecuteWithPrincipals(payload.NewPassword, principals); err != nil {
		return
	}

	// revoke old tokens
	now := timeNow()
	authInfo.TokenValidSince = &now
	if err = h.AuthInfoStore.UpdateAuth(&authInfo); err != nil {
		return
	}

	h.AuditTrail.Log(audit.Entry{
		AuthID: authInfo.ID,
		Event:  audit.EventResetPassword,
		Data: map[string]interface{}{
			"type": "forgot_password",
		},
	})

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: authInfo.ID,
	}, nil)

	resp = map[string]string{}

	return
}

func genericResetPasswordError() skyerr.Error {
	return skyerr.NewError(skyerr.ResourceNotFound, "user not found or code invalid")
}
