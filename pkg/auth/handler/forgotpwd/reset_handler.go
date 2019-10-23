package forgotpwd

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
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
	}).Methods("OPTIONS", "POST", "GET")
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
	return h.RequireAuthz(handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext), h)
}

type ForgotPasswordResetPayload struct {
	UserID       string `json:"user_id"`
	Code         string `json:"code"`
	ExpireAt     int64  `json:"expire_at"`
	ExpireAtTime time.Time
	NewPassword  string `json:"new_password"`
}

// nolint: gosec
// @JSONSchema
const ForgotPasswordResetRequestSchema = `
{
	"$id": "#ForgotPasswordResetRequest",
	"type": "object",
	"properties": {
		"user_id": { "type": "string" },
		"code": { "type": "string" },
		"expire_at": { "type": "string" },
		"new_password": { "type": "string" }
	}
}
`

func (payload ForgotPasswordResetPayload) Validate() error {
	// TODO(error): JSON schema
	if payload.UserID == "" {
		return skyerr.NewInvalid("empty user_id")
	}

	if payload.Code == "" {
		return skyerr.NewInvalid("empty code")
	}

	if payload.ExpireAt == 0 {
		return skyerr.NewInvalid("empty expire_at")
	}

	if payload.NewPassword == "" {
		return skyerr.NewInvalid("empty password")
	}

	return nil
}

/*
	@Operation POST /forgot_password/reset_password - Reset password
		Reset password using received recovery code.

		@Tag Forgot Password

		@RequestBody
			@JSONSchema {ForgotPasswordResetRequest}

		@Response 200 {EmptyResponse}

		@Callback password_update {PasswordUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type ForgotPasswordResetHandler struct {
	RequireAuthz         handler.RequireAuthz          `dependency:"RequireAuthz"`
	CodeGenerator        *forgotpwdemail.CodeGenerator `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker      *authAudit.PasswordChecker    `dependency:"PasswordChecker"`
	AuthInfoStore        authinfo.Store                `dependency:"AuthInfoStore"`
	UserProfileStore     userprofile.Store             `dependency:"UserProfileStore"`
	HookProvider         hook.Provider                 `dependency:"HookProvider"`
	PasswordAuthProvider password.Provider             `dependency:"PasswordAuthProvider"`
	AuditTrail           audit.Trail                   `dependency:"AuditTrail"`
	TxContext            db.TxContext                  `dependency:"TxContext"`
	Logger               *logrus.Entry                 `dependency:"HandlerLogger"`
	TaskQueue            async.Queue                   `dependency:"AsyncTaskQueue"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h ForgotPasswordResetHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

func (h ForgotPasswordResetHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ForgotPasswordResetHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := ForgotPasswordResetPayload{}
	if err := handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return nil, err
	}

	payload.ExpireAtTime = time.Unix(payload.ExpireAt, 0).UTC()

	return payload, nil
}

func (h ForgotPasswordResetHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ForgotPasswordResetPayload)

	err = passwordReseter{
		CodeGenerator:        h.CodeGenerator,
		PasswordChecker:      h.PasswordChecker,
		AuthInfoStore:        h.AuthInfoStore,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}.resetPassword(
		payload.UserID,
		payload.ExpireAtTime,
		payload.Code,
		payload.NewPassword,
	)
	if err != nil {
		return
	}

	var authInfo authinfo.AuthInfo
	err = h.AuthInfoStore.GetAuth(payload.UserID, &authInfo)
	if err != nil {
		return
	}

	userProfile, err := h.UserProfileStore.GetUserProfile(payload.UserID)
	if err != nil {
		return
	}

	user := model.NewUser(authInfo, userProfile)

	err = h.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonResetPassword,
			User:   user,
		},
		&user,
	)
	if err != nil {
		return
	}

	h.AuditTrail.Log(audit.Entry{
		UserID: user.ID,
		Event:  audit.EventResetPassword,
		Data: map[string]interface{}{
			"type": "forgot_password",
		},
	})

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: user.ID,
	}, nil)

	resp = map[string]string{}

	return
}
