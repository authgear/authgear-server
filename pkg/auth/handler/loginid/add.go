package loginid

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAddLoginIDHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/login_id/add").
		Handler(server.FactoryToHandler(&AddLoginIDHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type AddLoginIDHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f AddLoginIDHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AddLoginIDHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type AddLoginIDRequestPayload struct {
	LoginIDs []loginid.LoginID `json:"login_ids"`
}

// @JSONSchema
const AddLoginIDRequestSchema = `
{
	"$id": "#AddLoginIDRequest",
	"type": "object",
	"properties": {
		"login_ids": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"key": { "type": "string", "minLength": 1 },
					"value": { "type": "string", "minLength": 1 }
				},
				"required": ["key", "value"]
			},
			"minItems": 1
		}
	},
	"required": ["login_ids"]
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
	RequireAuthz             handler.RequireAuthz       `dependency:"RequireAuthz"`
	AuthInfoStore            authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider     password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider         principal.IdentityProvider `dependency:"IdentityProvider"`
	UserVerificationProvider userverify.Provider        `dependency:"UserVerificationProvider"`
	TxContext                db.TxContext               `dependency:"TxContext"`
	UserProfileStore         userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider             hook.Provider              `dependency:"HookProvider"`
}

func (h AddLoginIDHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
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

	err := db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

		principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(userID)
		if err != nil {
			return err
		}

		loginIDs := extractLoginIDs(principals)
		newLoginIDBeginIndex := len(loginIDs)
		loginIDs = append(loginIDs, payload.LoginIDs...)
		err = validateLoginIDs(h.PasswordAuthProvider, loginIDs, newLoginIDBeginIndex)
		if err != nil {
			if causes := validation.ErrorCauses(err); len(causes) > 0 {
				for i, cause := range causes {
					if cause.Pointer != "" {
						cause.Pointer = "/login_ids" + cause.Pointer
						causes[i] = cause
					}
				}
				return validation.NewValidationFailed("invalid login ID", causes)
			}
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

		for _, loginID := range payload.LoginIDs {
			newPrincipal, err := h.PasswordAuthProvider.MakePrincipal(userID, "", loginID, password.DefaultRealm)
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
		}

		err = h.UserVerificationProvider.UpdateVerificationState(authInfo, h.AuthInfoStore, principals)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}
