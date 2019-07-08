package sso

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/login", &LoginHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderName = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderName)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f LoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	AccessToken string `json:"access_token"`
}

// Validate request payload
func (p LoginRequestPayload) Validate() error {
	if p.AccessToken == "" {
		return skyerr.NewInvalidArgument("empty access token", []string{"access_token"})
	}

	return nil
}

// LoginHandler decodes code response and fetch access token from provider.
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/<provider>/link \
// <<EOF
// {
//     "token_response": {
//       "access_token": "<access_token>"
//     }
// }
// EOF
//
// {
//     "result": "OK"
// }
//
type LoginHandler struct {
	TxContext            db.TxContext               `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	OAuthAuthProvider    oauth.Provider             `dependency:"OAuthAuthProvider"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	ProviderFactory      *sso.ProviderFactory       `dependency:"SSOProviderFactory"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	Provider             sso.Provider
	ProviderName         string
}

func (h LoginHandler) WithTx() bool {
	return true
}

func (h LoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LoginRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	if err != nil {
		return payload, err
	}
	return payload, nil
}

func (h LoginHandler) Handle(req interface{}) (resp interface{}, err error) {
	if h.Provider == nil {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderName})
		return
	}

	payload := req.(LoginRequestPayload)
	oauthAuthInfo, err := h.Provider.GetAuthInfoByAccessTokenResp(sso.AccessTokenResp{
		AccessToken: payload.AccessToken,
		TokenType:   "Bearer",
	})
	if err != nil {
		return
	}

	handler := respHandler{
		TokenStore:           h.TokenStore,
		AuthInfoStore:        h.AuthInfoStore,
		OAuthAuthProvider:    h.OAuthAuthProvider,
		PasswordAuthProvider: h.PasswordAuthProvider,
		IdentityProvider:     h.IdentityProvider,
		UserProfileStore:     h.UserProfileStore,
		UserID:               oauthAuthInfo.State.UserID,
	}
	resp, err = handler.loginActionResp(oauthAuthInfo)

	return
}
