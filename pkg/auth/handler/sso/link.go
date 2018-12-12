package sso

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/skygeario/skygear-server/pkg/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachLinkHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/{provider}/link", &LinkHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type LinkHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LinkHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LinkHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderName = vars["provider"]
	h.Provider = h.ProviderFactory.NewProvider(h.ProviderName)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f LinkHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// LinkRequestPayload login handler request payload
type LinkRequestPayload sso.AccessTokenResp

// Validate request payload
func (p LinkRequestPayload) Validate() error {
	if p.AccessToken == "" {
		return skyerr.NewInvalidArgument("empty access token", []string{"access_token"})
	}

	return nil
}

// LinkHandler decodes code response and fetch access token from provider.
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
type LinkHandler struct {
	TxContext         db.TxContext           `dependency:"TxContext"`
	AuthContext       coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	OAuthAuthProvider oauth.Provider         `dependency:"OAuthAuthProvider"`
	AuthInfoStore     authinfo.Store         `dependency:"AuthInfoStore"`
	ProviderFactory   *sso.ProviderFactory   `dependency:"SSOProviderFactory"`
	Provider          sso.Provider
	ProviderName      string
}

func (h LinkHandler) WithTx() bool {
	return true
}

func (h LinkHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LinkRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	if err != nil {
		return payload, err
	}
	payload.Scope = strings.Split(payload.RawScope, " ")
	// some special handlings for facebook
	if payload.ExpiresIn == 0 && payload.RawExpires != 0 {
		payload.ExpiresIn = payload.RawExpires
	}
	if strings.ToLower(payload.TokenType) == "bearer" {
		payload.TokenType = "Bearer"
	}
	return payload, nil
}

func (h LinkHandler) Handle(req interface{}) (resp interface{}, err error) {
	if h.Provider == nil {
		err = skyerr.NewInvalidArgument("Provider is not supported", []string{h.ProviderName})
		return
	}

	payload := req.(LinkRequestPayload)
	oauthAuthInfo, err := h.Provider.GetAuthInfoByAccessTokenResp(sso.AccessTokenResp{
		AccessToken:  payload.AccessToken,
		TokenType:    payload.TokenType,
		ExpiresIn:    payload.ExpiresIn,
		Scope:        payload.Scope,
		RefreshToken: payload.RefreshToken,
	})
	if err != nil {
		return
	}

	// check if user is already linked
	userID := h.AuthContext.AuthInfo().ID
	_, err = h.OAuthAuthProvider.GetPrincipalByUserID(h.ProviderName, userID)
	if err == nil {
		err = skyerr.NewError(skyerr.InvalidArgument, "provider account already linked with existing user")
		return
	}

	if err != skydb.ErrUserNotFound {
		// some other error
		return
	}

	var info authinfo.AuthInfo
	if err = h.AuthInfoStore.GetAuth(userID, &info); err != nil {
		err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		return
	}

	now := timeNow()
	principal := oauth.NewPrincipal()
	principal.UserID = userID
	principal.ProviderName = oauthAuthInfo.ProviderName
	principal.ProviderUserID = oauthAuthInfo.ProviderUserID
	principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
	principal.UserProfile = oauthAuthInfo.ProviderUserProfile
	principal.CreatedAt = &now
	principal.UpdatedAt = &now
	err = h.OAuthAuthProvider.CreatePrincipal(principal)
	if err != nil {
		return
	}

	resp = "OK"

	return
}
