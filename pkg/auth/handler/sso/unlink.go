package sso

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	authprincipal "github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachUnlinkHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/unlink", &UnlinkHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type UnlinkHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f UnlinkHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UnlinkHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	return h.RequireAuthz(handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext), h)
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
	TxContext         db.TxContext                   `dependency:"TxContext"`
	AuthContext       coreAuth.ContextGetter         `dependency:"AuthContextGetter"`
	RequireAuthz      handler.RequireAuthz           `dependency:"RequireAuthz"`
	SessionProvider   session.Provider               `dependency:"SessionProvider"`
	OAuthAuthProvider oauth.Provider                 `dependency:"OAuthAuthProvider"`
	IdentityProvider  authprincipal.IdentityProvider `dependency:"IdentityProvider"`
	UserProfileStore  userprofile.Store              `dependency:"UserProfileStore"`
	HookProvider      hook.Provider                  `dependency:"HookProvider"`
	ProviderFactory   *sso.ProviderFactory           `dependency:"SSOProviderFactory"`
	ProviderID        string
}

func (h UnlinkHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h UnlinkHandler) WithTx() bool {
	return true
}

func (h UnlinkHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h UnlinkHandler) Handle(req interface{}) (resp interface{}, err error) {
	providerConfig, ok := h.ProviderFactory.GetProviderConfig(h.ProviderID)
	if !ok {
		err = skyerr.NewNotFound("unknown SSO provider")
		return
	}

	authInfo, _ := h.AuthContext.AuthInfo()
	sess, _ := h.AuthContext.Session()
	userID := authInfo.ID
	principal, err := h.OAuthAuthProvider.GetPrincipalByUser(oauth.GetByUserOptions{
		ProviderType: string(providerConfig.Type),
		ProviderKeys: oauth.ProviderKeysFromProviderConfig(providerConfig),
		UserID:       userID,
	})
	if err != nil {
		return
	}

	// principalID can be missing
	principalID := sess.PrincipalID
	if principalID != "" && principalID == principal.ID {
		err = authprincipal.ErrCurrentIdentityBeingDeleted
		return
	}

	err = h.OAuthAuthProvider.DeletePrincipal(principal)
	if err != nil {
		return
	}

	sessions, err := h.SessionProvider.List(userID)
	if err != nil {
		return
	}

	// filter sessions of deleted principal
	n := 0
	for _, session := range sessions {
		if session.PrincipalID == principal.ID {
			sessions[n] = session
			n++
		}
	}
	sessions = sessions[:n]

	err = h.SessionProvider.InvalidateBatch(sessions)
	if err != nil {
		return
	}

	var userProfile userprofile.UserProfile
	userProfile, err = h.UserProfileStore.GetUserProfile(userID)
	if err != nil {
		return
	}

	user := model.NewUser(*authInfo, userProfile)
	identity := model.NewIdentity(h.IdentityProvider, principal)
	err = h.HookProvider.DispatchEvent(
		event.IdentityDeleteEvent{
			User:     user,
			Identity: identity,
		},
		&user,
	)
	if err != nil {
		return
	}

	resp = map[string]string{}

	return
}
