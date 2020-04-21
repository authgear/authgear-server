package loginid

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

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

func AttachRemoveLoginIDHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/login_id/remove").
		Handler(server.FactoryToHandler(&RemoveLoginIDHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type RemoveLoginIDHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f RemoveLoginIDHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RemoveLoginIDHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type RemoveLoginIDRequestPayload struct {
	loginid.LoginID
}

// @JSONSchema
const RemoveLoginIDRequestSchema = `
{
	"$id": "#RemoveLoginIDRequest",
	"type": "object",
	"properties": {
		"key": { "type": "string", "minLength": 1 },
		"value": { "type": "string", "minLength": 1 }
	},
	"required": ["key", "value"]
}
`

/*
	@Operation POST /login_id/remove - Remove login ID
		Remove login ID from current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the login ID to remove.
			@JSONSchema {RemoveLoginIDRequest}

		@Response 200 {EmptyResponse}

		@Callback identity_delete {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type RemoveLoginIDHandler struct {
	Validator                *validation.Validator `dependency:"Validator"`
	RequireAuthz             handler.RequireAuthz  `dependency:"RequireAuthz"`
	AuthInfoStore            authinfo.Store        `dependency:"AuthInfoStore"`
	PasswordAuthProvider     password.Provider     `dependency:"PasswordAuthProvider"`
	UserVerificationProvider userverify.Provider   `dependency:"UserVerificationProvider"`
	TxContext                db.TxContext          `dependency:"TxContext"`
	UserProfileStore         userprofile.Store     `dependency:"UserProfileStore"`
	HookProvider             hook.Provider         `dependency:"HookProvider"`
	Logger                   *logrus.Entry         `dependency:"HandlerLogger"`
}

func (h RemoveLoginIDHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h RemoveLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h RemoveLoginIDHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	var payload RemoveLoginIDRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#RemoveLoginIDRequest", &payload); err != nil {
		return err
	}

	err := db.WithTx(h.TxContext, func() error {
		session := auth.GetSession(r.Context())
		userID := session.AuthnAttrs().UserID

		p, err := h.PasswordAuthProvider.GetPrincipalByLoginID(payload.Key, payload.Value)
		if err != nil {
			if errors.Is(err, principal.ErrNotFound) {
				err = password.ErrLoginIDNotFound
			}
			return err
		}
		if p.UserID != userID {
			return password.ErrLoginIDNotFound
		}

		err = h.PasswordAuthProvider.DeletePrincipal(p)
		if err != nil {
			return err
		}

		principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(userID)
		if err != nil {
			return err
		}
		err = validateLoginIDs(h.PasswordAuthProvider, extractLoginIDs(principals), -1)
		if err != nil {
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

		delete(authInfo.VerifyInfo, p.LoginID)
		err = h.UserVerificationProvider.UpdateVerificationState(authInfo, h.AuthInfoStore, principals)
		if err != nil {
			return err
		}

		identity := model.NewIdentity(p)
		err = h.HookProvider.DispatchEvent(
			event.IdentityDeleteEvent{
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
