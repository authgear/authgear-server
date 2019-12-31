package loginid

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
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

func AttachAddLoginIDHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/login_id/add", &AddLoginIDHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type AddLoginIDHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AddLoginIDHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AddLoginIDHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type AddLoginIDRequestPayload struct {
	password.LoginID
}

// @JSONSchema
const AddLoginIDRequestSchema = `
{
	"$id": "#AddLoginIDRequest",
	"type": "object",
	"properties": {
		"key": { "type": "string", "minLength": 1 },
		"value": { "type": "string", "minLength": 1 }
	},
	"required": ["key", "value"]
}
`

/*
	@Operation POST /login_id/add - Add login ID
		Add new login ID for current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the new login ID.
			@JSONSchema {AddLoginIDRequest}

		@Response 200 {EmptyResponse}

		@Callback identity_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AddLoginIDHandler struct {
	Validator                *validation.Validator      `dependency:"Validator"`
	AuthContext              coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz             handler.RequireAuthz       `dependency:"RequireAuthz"`
	AuthInfoStore            authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider     password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider         principal.IdentityProvider `dependency:"IdentityProvider"`
	UserVerificationProvider userverify.Provider        `dependency:"UserVerificationProvider"`
	TxContext                db.TxContext               `dependency:"TxContext"`
	UserProfileStore         userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider             hook.Provider              `dependency:"HookProvider"`
}

func (h AddLoginIDHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
}

func (h AddLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h AddLoginIDHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	var payload AddLoginIDRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#AddLoginIDRequest", &payload); err != nil {
		return err
	}

	err := hook.WithTx(h.HookProvider, h.TxContext, func() error {
		authInfo, _ := h.AuthContext.AuthInfo()
		userID := authInfo.ID

		principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(userID)
		if err != nil {
			return err
		}

		newLoginID := password.LoginID{Key: payload.Key, Value: payload.Value}
		loginIDs := extractLoginIDs(principals)
		newLoginIDIndex := len(loginIDs)
		loginIDs = append(loginIDs, newLoginID)
		err = validateLoginIDs(h.PasswordAuthProvider, loginIDs, newLoginIDIndex)
		if err != nil {
			return err
		}

		newPrincipal, err := h.PasswordAuthProvider.MakePrincipal(userID, "", newLoginID, password.DefaultRealm)
		if err != nil {
			return err
		}
		if len(principals) > 0 {
			newPrincipal.HashedPassword = principals[0].HashedPassword
		} else {
			// NOTE: if there is no existing password principals,
			// we use a empty password hash to make it unable to be used
			// to login.
			newPrincipal.HashedPassword = nil
		}

		err = h.PasswordAuthProvider.CreatePrincipal(newPrincipal)
		if err != nil {
			return err
		}
		principals = append(principals, newPrincipal)

		var userProfile userprofile.UserProfile
		userProfile, err = h.UserProfileStore.GetUserProfile(userID)
		if err != nil {
			return err
		}
		user := model.NewUser(*authInfo, userProfile)

		err = h.UserVerificationProvider.UpdateVerificationState(authInfo, h.AuthInfoStore, principals)
		if err != nil {
			return err
		}

		identity := model.NewIdentity(h.IdentityProvider, newPrincipal)
		err = h.HookProvider.DispatchEvent(
			event.IdentityCreateEvent{
				User:     user,
				Identity: identity,
			},
			&user,
		)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}
