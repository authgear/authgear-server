package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachChangePasswordHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/change_password").
		Handler(pkg.MakeHandler(authDependency, newChangePasswordHandler)).
		Methods("OPTIONS", "POST")
}

type PasswordFlow interface {
	ChangePassword(userID string, OldPassword string, newPassword string) error
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
	Validator        *validation.Validator
	AuthInfoStore    authinfo.Store
	TxContext        db.TxContext
	UserProfileStore userprofile.Store
	HookProvider     hook.Provider
	TaskQueue        async.Queue
	Interactions     PasswordFlow
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h ChangePasswordHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
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
		sess := auth.GetSession(r.Context())
		userID := sess.AuthnAttrs().UserID

		if err := h.Interactions.ChangePassword(
			userID, payload.OldPassword, payload.NewPassword,
		); err != nil {
			return err
		}

		authInfo := &authinfo.AuthInfo{}
		if err := h.AuthInfoStore.GetAuth(userID, authInfo); err != nil {
			return err
		}

		userProfile, err := h.UserProfileStore.GetUserProfile(userID)
		if err != nil {
			return err
		}

		user := model.NewUser(*authInfo, userProfile)
		identity := model.NewIdentityFromAttrs(sess.AuthnAttrs())

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
				AuthID: userID,
			},
		})

		return nil
	})
	return
}
