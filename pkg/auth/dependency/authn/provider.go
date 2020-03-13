package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type SignupProvider interface {
	// CreateUserWithLoginIDs is sign up.
	CreateUserWithLoginIDs(
		loginIDs []loginid.LoginID,
		plainPassword string,
		metadata map[string]interface{},
		onUserDuplicate model.OnUserDuplicate,
	) (authInfo *authinfo.AuthInfo, userprofile *userprofile.UserProfile, firstPrincipal principal.Principal, tasks []async.TaskSpec, err error)
}

type LoginProvider interface {
	// AuthenticateWithLoginID is sign in.
	AuthenticateWithLoginID(loginID loginid.LoginID, plainPassword string) (authInfo *authinfo.AuthInfo, principal principal.Principal, err error)
}

type OAuthProvider interface {
	// AuthenticateWithOAuth is oauth sign up/sign in.
	AuthenticateWithOAuth(oauthAuthInfo sso.AuthInfo, codeChallenge string, loginState sso.LoginState) (code *sso.SkygearAuthorizationCode, tasks []async.TaskSpec, err error)

	// LinkOAuth links oauth identity with an existing user.
	LinkOAuth(oauthAuthInfo sso.AuthInfo, codeChallenge string, linkState sso.LinkState) (code *sso.SkygearAuthorizationCode, err error)

	ExtractAuthorizationCode(code *sso.SkygearAuthorizationCode) (authInfo *authinfo.AuthInfo, userProfile *userprofile.UserProfile, prin principal.Principal, err error)
}
