package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachChangePasswordHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/change_password").
		Handler(server.FactoryToHandler(&ChangePasswordHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

// ChangePasswordHandlerFactory creates ChangePasswordHandler
type ChangePasswordHandlerFactory struct {
	Dependency pkg.DependencyMap
}

// NewHandler creates new handler
func (f ChangePasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ChangePasswordHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
		"password": { "type": "string", "minLength": 1 },
		"old_password": { "type": "string", "minLength": 1 }
	},
	"required": ["password", "old_password"]
}
`

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
	Validator            *validation.Validator      `dependency:"Validator"`
	RequireAuthz         handler.RequireAuthz       `dependency:"RequireAuthz"`
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
	return policy.RequireValidUser
}

func (h ChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	result, err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h ChangePasswordHandler) Handle(w http.ResponseWriter, r *http.Request) (resp model.AuthResponse, err error) {
	var payload ChangePasswordRequestPayload
	if err = handler.BindJSONBody(r, w, h.Validator, "#ChangePasswordRequest", &payload); err != nil {
		return
	}

	err = db.WithTx(h.TxContext, func() error {
		authinfo := auth.GetAuthInfo(r.Context())
		sess := auth.GetSession(r.Context())

		if err := h.PasswordChecker.ValidatePassword(authAudit.ValidatePasswordPayload{
			PlainPassword: payload.NewPassword,
			AuthID:        authinfo.ID,
		}); err != nil {
			return err
		}

		principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authinfo.ID)
		if err != nil {
			return err
		}
		if len(principals) == 0 {
			err = skyerr.NewInvalid("user has no password")
			return err
		}

		principal := principals[0]
		for _, p := range principals {
			if p.ID == sess.AuthnAttrs().PrincipalID {
				principal = p
			}
			err = p.VerifyPassword(payload.OldPassword)
			if err != nil {
				return err
			}
			err = h.PasswordAuthProvider.UpdatePassword(p, payload.NewPassword)
			if err != nil {
				return err
			}
		}

		// Get Profile
		userProfile, err := h.UserProfileStore.GetUserProfile(authinfo.ID)
		if err != nil {
			return err
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
			return err
		}

		resp = model.NewAuthResponseWithUserIdentity(user, identity)

		// password house keeper
		h.TaskQueue.Enqueue(async.TaskSpec{
			Name: task.PwHousekeeperTaskName,
			Param: task.PwHousekeeperTaskParam{
				AuthID: authinfo.ID,
			},
		})

		return nil
	})
	return
}
