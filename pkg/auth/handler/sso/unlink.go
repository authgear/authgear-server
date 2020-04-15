package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var errOauthIdentityNotFound = skyerr.NotFound.WithReason("OAuthIdentityNotFound").New("oauth identity not found")

func AttachUnlinkHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/unlink").
		Handler(pkg.MakeHandler(authDependency, newUnlinkHandler)).
		Methods("OPTIONS", "POST")
}

type unlinkSessionManager interface {
	List(userID string) ([]auth.AuthSession, error)
	Revoke(auth.AuthSession) error
}

/*
	@Operation POST /sso/{provider_id}/unlink - Unlink SSO provider
		Unlink the specified SSO provider from the current user.

		@Tag SSO
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Parameter {SSOProviderID}
		@Response 200 {EmptyResponse}

		@Callback identity_delete {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type UnlinkHandler struct {
	TxContext         db.TxContext
	RequireAuthz      handler.RequireAuthz
	SessionManager    unlinkSessionManager
	OAuthAuthProvider oauth.Provider
	AuthInfoStore     authinfo.Store
	UserProfileStore  userprofile.Store
	HookProvider      hook.Provider
	ProviderFactory   *sso.OAuthProviderFactory
}

func (h UnlinkHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h UnlinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		response.Error = err
	} else {
		result, err := h.Handle(r)
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	}
	handler.WriteResponse(w, response)
}

func (h UnlinkHandler) Handle(r *http.Request) (resp interface{}, err error) {
	err = db.WithTx(h.TxContext, func() error {
		vars := mux.Vars(r)
		providerID := vars["provider"]

		providerConfig, ok := h.ProviderFactory.GetOAuthProviderConfig(providerID)
		if !ok {
			return skyerr.NewNotFound("unknown SSO provider")
		}

		sess := auth.GetSession(r.Context())
		userID := sess.AuthnAttrs().UserID
		prin, err := h.OAuthAuthProvider.GetPrincipalByUser(oauth.GetByUserOptions{
			ProviderType: string(providerConfig.Type),
			ProviderKeys: oauth.ProviderKeysFromProviderConfig(providerConfig),
			UserID:       userID,
		})
		if err != nil {
			if errors.Is(err, principal.ErrNotFound) {
				err = errOauthIdentityNotFound
			}
			return err
		}

		// principalID can be missing
		principalID := sess.AuthnAttrs().PrincipalID
		if principalID != "" && principalID == prin.ID {
			err = principal.ErrCurrentIdentityBeingDeleted
			return err
		}

		err = h.OAuthAuthProvider.DeletePrincipal(prin)
		if err != nil {
			return err
		}

		sessions, err := h.SessionManager.List(userID)
		if err != nil {
			return err
		}

		// delete sessions of deleted principal
		for _, session := range sessions {
			if session.AuthnAttrs().PrincipalID != prin.ID {
				continue
			}

			err := h.SessionManager.Revoke(session)
			if err != nil {
				return err
			}
		}

		authInfo := &authinfo.AuthInfo{}
		if err := h.AuthInfoStore.GetAuth(userID, authInfo); err != nil {
			return err
		}

		var userProfile userprofile.UserProfile
		userProfile, err = h.UserProfileStore.GetUserProfile(userID)
		if err != nil {
			return err
		}

		user := model.NewUser(*authInfo, userProfile)
		identity := model.NewIdentity(prin)
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

		resp = struct{}{}
		return nil
	})
	return
}
