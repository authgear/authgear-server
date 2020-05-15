package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/auth"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachResetPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/reset_password").
		Handler(auth.MakeHandler(authDependency, newResetPasswordHandler)).
		Methods("OPTIONS", "POST")
}

type ResetPasswordFlow interface {
	ResetPassword(userID string, password string) error
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
		"user_id": { "type": "string" },
		"password": { "type": "string" }
	},
	"required": ["user_id", "password"]
}
`

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
	Validator        *validation.Validator `dependency:"Validator"`
	UserProfileStore userprofile.Store     `dependency:"UserProfileStore"`
	AuthInfoStore    authinfo.Store        `dependency:"AuthInfoStore"`
	TxContext        db.TxContext          `dependency:"TxContext"`
	TaskQueue        async.Queue           `dependency:"AsyncTaskQueue"`
	HookProvider     hook.Provider         `dependency:"HookProvider"`
	Interactions     ResetPasswordFlow
}

func (h ResetPasswordHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireMasterKey)
}

func (h ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h ResetPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ResetPasswordRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#ResetPasswordRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		if err := h.Interactions.ResetPassword(payload.UserID, payload.Password); err != nil {
			return err
		}

		authinfo := authinfo.AuthInfo{}
		if err := h.AuthInfoStore.GetAuth(payload.UserID, &authinfo); err != nil {
			return err
		}

		var profile userprofile.UserProfile
		if profile, err = h.UserProfileStore.GetUserProfile(authinfo.ID); err != nil {
			return err
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
			return err
		}

		// password house keeper
		h.TaskQueue.Enqueue(async.TaskSpec{
			Name: task.PwHousekeeperTaskName,
			Param: task.PwHousekeeperTaskParam{
				AuthID: authinfo.ID,
			},
		})

		resp = struct{}{}
		return nil
	})
	return
}
