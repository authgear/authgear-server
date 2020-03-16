package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	authModel "github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// AttachSetDisableHandler attaches SetDisableHandler to server
func AttachSetDisableHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/disable/set").
		Handler(server.FactoryToHandler(&SetDisableHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

// SetDisableHandlerFactory creates SetDisableHandler
type SetDisableHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new SetDisableHandler
func (f SetDisableHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SetDisableHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type setDisableUserPayload struct {
	UserID       string `json:"user_id"`
	Disabled     bool   `json:"disabled"`
	Message      string `json:"message"`
	ExpiryString string `json:"expiry"`
	expiry       *time.Time
}

func (p *setDisableUserPayload) SetDefaultValue() {
	if p.ExpiryString != "" {
		expiry, err := time.Parse(time.RFC3339, p.ExpiryString)
		if err != nil {
			// should be already validated, so panic if unexpected.
			panic(err)
		}
		p.expiry = &expiry
	}
}

// @JSONSchema
const SetDisableRequestSchema = `
{
	"$id": "#SetDisableRequest",
	"type": "object",
	"properties": {
		"user_id": { "type": "string", "minLength": 1 },
		"disabled": { "type": "boolean" },
		"message": { "type": "string", "minLength": 1 },
		"expiry": { "type": "string", "format": "date-time" }
	},
	"required": ["user_id", "disabled"]
}
`

/*
	@Operation POST /disable/set - Set user disabled status
		Disable/enable target user.

		@Tag Administration
		@SecurityRequirement master_key
		@SecurityRequirement access_token

		@RequestBody
			Describe target user and desired disable status.
			@JSONSchema {SetDisableRequest}
			@JSONExample EnableUser - Enable user
				{
					"auth_id": "F1D4AAAC-A31A-4471-92B2-6E08376BDD87",
					"disabled": false
				}
			@JSONExample DisableUser - Disable user permanently
				{
					"auth_id": "F1D4AAAC-A31A-4471-92B2-6E08376BDD87",
					"disabled": true
				}
			@JSONExample DisableUserExpiry - Disable user with expiry
				{
					"auth_id": "F1D4AAAC-A31A-4471-92B2-6E08376BDD87",
					"disabled": true,
					"message": "Banned",
					"expiry": "2019-07-31T09:39:22.349Z"
				}

		@Response 200 {EmptyResponse}

		@Callback user_update {UserUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type SetDisableHandler struct {
	Validator        *validation.Validator  `dependency:"Validator"`
	AuthContext      coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz   `dependency:"RequireAuthz"`
	AuthInfoStore    authinfo.Store         `dependency:"AuthInfoStore"`
	UserProfileStore userprofile.Store      `dependency:"UserProfileStore"`
	HookProvider     hook.Provider          `dependency:"HookProvider"`
	TxContext        db.TxContext           `dependency:"TxContext"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h SetDisableHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireMasterKey)
}

func (h SetDisableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h SetDisableHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload setDisableUserPayload
	if err = handler.BindJSONBody(r, w, h.Validator, "#SetDisableRequest", &payload); err != nil {
		return
	}

	h.TxContext.UseHook(h.HookProvider)
	err = db.WithTx(h.TxContext, func() error {
		info := authinfo.AuthInfo{}
		if err = h.AuthInfoStore.GetAuth(payload.UserID, &info); err != nil {
			return err
		}

		profile, err := h.UserProfileStore.GetUserProfile(info.ID)
		if err != nil {
			return err
		}

		oldUser := authModel.NewUser(info, profile)

		info.Disabled = payload.Disabled
		if !info.Disabled {
			info.DisabledMessage = ""
			info.DisabledExpiry = nil
		} else {
			info.DisabledMessage = payload.Message
			info.DisabledExpiry = payload.expiry
		}

		if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
			return err
		}

		user := authModel.NewUser(info, profile)

		err = h.HookProvider.DispatchEvent(
			event.UserUpdateEvent{
				Reason:     event.UserUpdateReasonAdministrative,
				User:       oldUser,
				IsDisabled: &payload.Disabled,
			},
			&user,
		)
		if err != nil {
			return err
		}

		resp = struct{}{}
		return nil
	})
	return
}
