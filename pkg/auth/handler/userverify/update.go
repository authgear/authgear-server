package userverify

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
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

func AttachUpdateHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/update_verify_state", &UpdateHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type UpdateHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f UpdateHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UpdateHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type UpdateVerifyStatePayload struct {
	UserID           string          `json:"user_id"`
	VerifyInfo       map[string]bool `json:"verify_info"`
	ManuallyVerified *bool           `json:"is_manually_verified"`
}

// @JSONSchema
const UpdateVerifyStateRequestSchema = `
{
	"$id": "#UpdateVerifyStateRequest",
	"type": "object",
	"properties": {
		"user_id": { "type": "string", "minLength": 1 },
		"verify_info": { "type": "object" },
		"is_manually_verified": { "type": "boolean" }
	},
	"required": ["user_id"]
}
`

/*
	@Operation POST /update_verify_state - Update user verification state
		Update verification state of the target user.

		@Tag Administration
		@SecurityRequirement master_key
		@SecurityRequirement access_token

		@RequestBody
			Describe target user and desired verification state.
			@JSONSchema {UpdateVerifyStateRequest}

		@Response 200 {UserResponse}

		@Callback user_update {UserUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type UpdateHandler struct {
	Validator                *validation.Validator  `dependency:"Validator"`
	AuthContext              coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	RequireAuthz             handler.RequireAuthz   `dependency:"RequireAuthz"`
	AuthInfoStore            authinfo.Store         `dependency:"AuthInfoStore"`
	UserProfileStore         userprofile.Store      `dependency:"UserProfileStore"`
	PasswordAuthProvider     password.Provider      `dependency:"PasswordAuthProvider"`
	UserVerificationProvider userverify.Provider    `dependency:"UserVerificationProvider"`
	HookProvider             hook.Provider          `dependency:"HookProvider"`
	TxContext                db.TxContext           `dependency:"TxContext"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h UpdateHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload UpdateVerifyStatePayload
	if err = handler.BindJSONBody(r, w, h.Validator, "#UpdateVerifyStateRequest", &payload); err != nil {
		return
	}

	err = hook.WithTx(h.HookProvider, h.TxContext, func() error {
		info := authinfo.AuthInfo{}
		if err = h.AuthInfoStore.GetAuth(payload.UserID, &info); err != nil {
			return err
		}

		profile, err := h.UserProfileStore.GetUserProfile(info.ID)
		if err != nil {
			return err
		}

		oldUser := authModel.NewUser(info, profile)

		principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(info.ID)
		if err != nil {
			return err
		}

		if payload.VerifyInfo != nil {
			verifyInfo := map[string]bool{}
			for loginID, verified := range payload.VerifyInfo {
				if verified {
					verifyInfo[loginID] = true
				}
			}
			info.VerifyInfo = verifyInfo
		}
		if payload.ManuallyVerified != nil {
			info.ManuallyVerified = *payload.ManuallyVerified
		}

		err = h.UserVerificationProvider.UpdateVerificationState(&info, h.AuthInfoStore, principals)
		if err != nil {
			return err
		}

		user := authModel.NewUser(info, profile)

		isVerified := info.IsVerified()
		err = h.HookProvider.DispatchEvent(
			event.UserUpdateEvent{
				Reason:     event.UserUpdateReasonAdministrative,
				User:       oldUser,
				VerifyInfo: &info.VerifyInfo,
				IsVerified: &isVerified,
			},
			&user,
		)
		if err != nil {
			return err
		}

		resp = authModel.NewAuthResponseWithUser(user)
		return nil
	})
	return
}
