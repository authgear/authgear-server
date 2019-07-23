package sso

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
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
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f UnlinkHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// UnlinkHandler decodes code response and fetch access token from provider.
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/<provider>/unlink \
// <<EOF
//
// {
//     "result": {}
// }
//
type UnlinkHandler struct {
	TxContext         db.TxContext               `dependency:"TxContext"`
	AuthContext       coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	OAuthAuthProvider oauth.Provider             `dependency:"OAuthAuthProvider"`
	IdentityProvider  principal.IdentityProvider `dependency:"IdentityProvider"`
	UserProfileStore  userprofile.Store          `dependency:"UserProfileStore"`
	HookProvider      hook.Provider              `dependency:"HookProvider"`
	ProviderFactory   *sso.ProviderFactory       `dependency:"SSOProviderFactory"`
	ProviderID        string
}

func (h UnlinkHandler) WithTx() bool {
	return true
}

func (h UnlinkHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

func (h UnlinkHandler) Handle(req interface{}) (resp interface{}, err error) {
	providerConfig, ok := h.ProviderFactory.GetProviderConfig(h.ProviderID)
	if !ok {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderID})
		return
	}

	userID := h.AuthContext.AuthInfo().ID
	principal, err := h.OAuthAuthProvider.GetPrincipalByUser(oauth.GetByUserOptions{
		ProviderType: string(providerConfig.Type),
		ProviderKeys: oauth.ProviderKeysFromProviderConfig(providerConfig),
		UserID:       userID,
	})
	if err != nil {
		return
	}

	err = h.OAuthAuthProvider.DeletePrincipal(principal)
	if err != nil {
		return
	}

	var userProfile userprofile.UserProfile
	userProfile, err = h.UserProfileStore.GetUserProfile(userID)
	if err != nil {
		return
	}

	user := model.NewUser(*h.AuthContext.AuthInfo(), userProfile)
	identity := model.NewIdentity(h.IdentityProvider, principal)
	err = h.HookProvider.DispatchEvent(
		event.IdentityDeleteEvent{
			User:     &user,
			Identity: &identity,
		},
		&user,
	)
	if err != nil {
		return
	}

	resp = map[string]string{}

	return
}
