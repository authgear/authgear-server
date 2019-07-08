package sso

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
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
	h.ProviderName = vars["provider"]
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
//     "result": "OK"
// }
//
type UnlinkHandler struct {
	TxContext         db.TxContext           `dependency:"TxContext"`
	AuthContext       coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	OAuthAuthProvider oauth.Provider         `dependency:"OAuthAuthProvider"`
	ProviderName      string
}

func (h UnlinkHandler) WithTx() bool {
	return true
}

func (h UnlinkHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

func (h UnlinkHandler) Handle(req interface{}) (resp interface{}, err error) {
	userID := h.AuthContext.AuthInfo().ID
	principal, err := h.OAuthAuthProvider.GetPrincipalByUserID(h.ProviderName, userID)
	if err != nil {
		return
	}

	err = h.OAuthAuthProvider.DeletePrincipal(principal)
	if err != nil {
		return
	}

	resp = "OK"

	return
}
