package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachResetPasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/reset_password", &ResetPasswordHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type ResetPasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ResetPasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ResetPasswordHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext), h.AuthContext, h)
}

type ResetPasswordRequestPayload struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

// nolint: gosec
// @JSONSchema
const ResetPasswordRequestSchema = `
{
	"$id": "#ResetPasswordRequest",
	"type": "object",
	"properties": {
		"auth_id": { "type": "string" },
		"password": { "type": "string" }
	}
}
`

func (p ResetPasswordRequestPayload) Validate() error {
	if p.UserID == "" {
		return skyerr.NewInvalidArgument("invalid user id", []string{"user_id"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

/*
	@Operation POST /reset_password - Reset user password
		Reset password of target user.

		@Tag Administration
		@SecurityRequirement master_key
		@SecurityRequirement access_token

		@RequestBody
			Describe target user and new password.
			@JSONSchema {ResetPasswordRequest}

		@Response 200 {EmptyResponse}

		@Callback password_update {PasswordUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type ResetPasswordHandler struct {
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	PasswordChecker      *authAudit.PasswordChecker `dependency:"PasswordChecker"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	AuditTrail           audit.Trail                `dependency:"AuditTrail"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	TaskQueue            async.Queue                `dependency:"AsyncTaskQueue"`
	HookProvider         hook.Provider              `dependency:"HookProvider"`
}

func (h ResetPasswordHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h ResetPasswordHandler) WithTx() bool {
	return true
}

func (h ResetPasswordHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ResetPasswordRequestPayload{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h ResetPasswordHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ResetPasswordRequestPayload)

	authinfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(payload.UserID, &authinfo); e != nil {
		if err == skydb.ErrUserNotFound {
			// logger.Info("Auth info not found when setting disabled user status")
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
		// logger.WithError(err).Error("Unable to get auth info when setting disabled user status")
		err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
		return
	}

	resetPwdCtx := password.ResetPasswordRequestContext{
		PasswordChecker:      h.PasswordChecker,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}

	if err = resetPwdCtx.ExecuteWithUserID(payload.Password, authinfo.ID); err != nil {
		return
	}

	var profile userprofile.UserProfile
	if profile, err = h.UserProfileStore.GetUserProfile(authinfo.ID); err != nil {
		return
	}

	user := model.NewUser(authinfo, profile)

	err = h.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonAdministrative,
			User:   user,
		},
		&user,
	)
	if err != nil {
		return
	}

	h.AuditTrail.Log(audit.Entry{
		UserID: authinfo.ID,
		Event:  audit.EventResetPassword,
	})

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: authinfo.ID,
	}, nil)

	resp = map[string]string{}

	return
}
