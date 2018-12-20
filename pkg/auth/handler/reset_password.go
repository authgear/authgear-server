package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/response"
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
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f ResetPasswordHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		policy.AnyOf(
			authz.PolicyFunc(policy.RequireAdminRole),
			authz.PolicyFunc(policy.RequireMasterKey),
		),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type ResetPasswordRequestPayload struct {
	AuthInfoID string `json:"auth_id"`
	Password   string `json:"password"`
}

func (p ResetPasswordRequestPayload) Validate() error {
	if p.AuthInfoID == "" {
		return skyerr.NewInvalidArgument("invalid auth id", []string{"auth_id"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

// ResetPasswordHandler handles signup request
type ResetPasswordHandler struct {
	PasswordChecker      *authAudit.PasswordChecker `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	AuditTrail           audit.Trail                `dependency:"AuditTrail"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	TaskQueue            async.Queue                `dependency:"AsyncTaskQueue"`
}

func (h ResetPasswordHandler) WithTx() bool {
	return true
}

func (h ResetPasswordHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ResetPasswordRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h ResetPasswordHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ResetPasswordRequestPayload)

	authinfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(payload.AuthInfoID, &authinfo); e != nil {
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

	now := timeNow()
	authinfo.TokenValidSince = &now
	err = h.AuthInfoStore.UpdateAuth(&authinfo)
	if err != nil {
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(authinfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authinfo.ID); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	resp = response.NewAuthResponse(authinfo, userProfile, token.AccessToken)
	h.AuditTrail.Log(audit.Entry{
		AuthID: authinfo.ID,
		Event:  audit.EventResetPassword,
	})

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: authinfo.ID,
	}, nil)

	return
}
