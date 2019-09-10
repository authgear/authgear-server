package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/session"

	"github.com/skygeario/skygear-server/pkg/auth/event"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
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

func AttachChangePasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/change_password", &ChangePasswordHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ChangePasswordHandlerFactory creates ChangePasswordHandler
type ChangePasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new handler
func (f ChangePasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ChangePasswordHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(h, h.AuthContext, h)
}

type ChangePasswordRequestPayload struct {
	NewPassword string `json:"password"`
	OldPassword string `json:"old_password"`
}

// nolint:gosec
// @JSONSchema
const ChangePasswordRequestSchema = `
{
	"$id": "#ChangePasswordRequest",
	"type": "object",
	"properties": {
		"password": { "type": "string" },
		"old_password": { "type": "string" }
	}
}
`

func (p ChangePasswordRequestPayload) Validate() error {
	if p.OldPassword == "" {
		return skyerr.NewInvalidArgument("empty old password", []string{"old_password"})
	}
	if p.NewPassword == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}
	return nil
}

/*
	@Operation POST /change_password - Change password
		Changes current user password.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe old and new password.
			@JSONSchema {ChangePasswordRequest}

		@Response 200
			Return user and new access token.
			@JSONSchema {AuthResponse}

		@Callback password_update {PasswordUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type ChangePasswordHandler struct {
	AuditTrail           audit.Trail                `dependency:"AuditTrail"`
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
	PasswordChecker      *authAudit.PasswordChecker `dependency:"PasswordChecker"`
	SessionProvider      session.Provider           `dependency:"SessionProvider"`
	SessionWriter        session.Writer             `dependency:"SessionWriter"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider         hook.Provider              `dependency:"HookProvider"`
	TaskQueue            async.Queue                `dependency:"AsyncTaskQueue"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h ChangePasswordHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h ChangePasswordHandler) WithTx() bool {
	return true
}

// DecodeRequest decode the request payload
func (h ChangePasswordHandler) DecodeRequest(request *http.Request) (payload ChangePasswordRequestPayload, err error) {
	err = handler.DecodeJSONBody(request, &payload)
	return
}

func (h ChangePasswordHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error
	var result interface{}
	defer func() {
		if err == nil {
			h.HookProvider.DidCommitTx()
			authResp := result.(model.AuthResponse)
			h.SessionWriter.WriteSession(resp, &authResp.AccessToken, nil)
			handler.WriteResponse(resp, handler.APIResponse{Result: authResp})
		} else {
			handler.WriteResponse(resp, handler.APIResponse{Err: skyerr.MakeError(err)})
		}
	}()

	payload, err := h.DecodeRequest(req)
	if err != nil {
		return
	}

	if err = payload.Validate(); err != nil {
		return
	}

	result, err = handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		result, err = h.Handle(payload)
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
}

func (h ChangePasswordHandler) Handle(payload ChangePasswordRequestPayload) (resp model.AuthResponse, err error) {
	authinfo := h.AuthContext.AuthInfo()

	if err = h.PasswordChecker.ValidatePassword(authAudit.ValidatePasswordPayload{
		PlainPassword: payload.NewPassword,
		AuthID:        authinfo.ID,
	}); err != nil {
		return
	}

	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authinfo.ID)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		return
	}

	principal := principals[0]
	for _, p := range principals {
		if p.ID == h.AuthContext.Session().PrincipalID {
			principal = p
		}
		if !p.IsSamePassword(payload.OldPassword) {
			err = skyerr.NewError(skyerr.InvalidCredentials, "Incorrect old password")
			return
		}
		err = h.PasswordAuthProvider.UpdatePassword(p, payload.NewPassword)
		if err != nil {
			return
		}
	}

	// refresh session
	session := h.AuthContext.Session()
	err = h.SessionProvider.Refresh(h.AuthContext.Session())
	if err != nil {
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

	user := model.NewUser(*authinfo, userProfile)
	identity := model.NewIdentity(h.IdentityProvider, principal)

	err = h.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonChangePassword,
			User:   user,
		},
		&user,
	)
	if err != nil {
		return
	}

	resp = model.NewAuthResponse(user, identity, session, "")

	h.AuditTrail.Log(audit.Entry{
		UserID: authinfo.ID,
		Event:  audit.EventChangePassword,
	})

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: authinfo.ID,
	}, nil)

	return
}
