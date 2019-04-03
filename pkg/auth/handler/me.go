package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachMeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/me", &MeHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type MeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f MeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &MeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f MeHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// MeHandler handles method of the me request, responds with current user data.
//
// The handler also:
// 1. refresh access token with a newly generated one
// 2. populate the activity time to user
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/me <<EOF
//  {
//  }
//  EOF
//
// {
//   "user_id": "3df4b52b-bd58-4fa2-8aee-3d44fd7f974d",
//   "last_login_at": "2016-09-08T06:42:59.871181Z",
//   "last_seen_at": "2016-09-08T07:15:18.026567355Z",
// }
type MeHandler struct {
	AuthContext          coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext            db.TxContext           `dependency:"TxContext"`
	TokenStore           authtoken.Store        `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store         `dependency:"AuthInfoStore"`
	UserProfileStore     userprofile.Store      `dependency:"UserProfileStore"`
	PasswordAuthProvider password.Provider      `dependency:"PasswordAuthProvider"`
}

func (h MeHandler) WithTx() bool {
	return true
}

func (h MeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

func (h MeHandler) Handle(req interface{}) (resp interface{}, err error) {
	authInfo := h.AuthContext.AuthInfo()

	token, err := h.TokenStore.NewToken(authInfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	authResp := response.NewAuthResponse(*authInfo, userProfile, token.AccessToken)
	// Get all loginIDs
	if principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID); err == nil {
		loginIDs := password.PrincipalsToLoginIDs(principals)
		authResp.LoginIDs = loginIDs
	}

	now := timeNow()
	authInfo.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(authInfo); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	return authResp, nil
}
