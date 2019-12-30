package loginid

import (
	"errors"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachRemoveLoginIDHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/login_id/remove", &RemoveLoginIDHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RemoveLoginIDHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RemoveLoginIDHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RemoveLoginIDHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type RemoveLoginIDRequestPayload struct {
	password.LoginID
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
	Validator            *validation.Validator      `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz       `dependency:"RequireAuthz"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
	SessionProvider      session.Provider           `dependency:"SessionProvider"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider         hook.Provider              `dependency:"HookProvider"`
}

func (h RemoveLoginIDHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
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

	err := hook.WithTx(h.HookProvider, h.TxContext, func() error {
		authInfo, _ := h.AuthContext.AuthInfo()
		session, _ := h.AuthContext.Session()
		userID := authInfo.ID

		var p password.Principal
		err := h.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm(payload.Key, payload.Value, password.DefaultRealm, &p)
		if err != nil {
			if errors.Is(err, principal.ErrNotFound) {
				err = password.ErrLoginIDNotFound
			}
			return err
		}
		if p.UserID != userID {
			return password.ErrLoginIDNotFound
		}

		if session.PrincipalID != "" && session.PrincipalID == p.ID {
			err = principal.ErrCurrentIdentityBeingDeleted
			return err
		}

		err = h.PasswordAuthProvider.DeletePrincipal(&p)
		if err != nil {
			return err
		}

		principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(userID)
		if err != nil {
			return err
		}
		err = validateLoginIDs(h.PasswordAuthProvider, extractLoginIDs(principals), nil)
		if err != nil {
			return err
		}

		sessions, err := h.SessionProvider.List(userID)
		if err != nil {
			return err
		}

		// filter sessions of deleted principal
		n := 0
		for _, session := range sessions {
			if session.PrincipalID == p.ID {
				sessions[n] = session
				n++
			}
		}
		sessions = sessions[:n]

		err = h.SessionProvider.InvalidateBatch(sessions)
		if err != nil {
			return err
		}

		var userProfile userprofile.UserProfile
		userProfile, err = h.UserProfileStore.GetUserProfile(userID)
		if err != nil {
			return err
		}

		user := model.NewUser(*authInfo, userProfile)
		identity := model.NewIdentity(h.IdentityProvider, &p)
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
