package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type SignupProvider interface {
	// CreateUserWithLoginIDs is sign up.
	CreateUserWithLoginIDs(
		loginIDs []loginid.LoginID,
		plainPassword string,
		metadata map[string]interface{},
		onUserDuplicate model.OnUserDuplicate,
	) (authInfo *authinfo.AuthInfo, userprofile *userprofile.UserProfile, firstPrincipal principal.Principal, err error)
}

type LoginProvider interface {
	// AuthenticateWithLoginID is sign in.
	AuthenticateWithLoginID(loginID loginid.LoginID, plainPassword string) (authInfo *authinfo.AuthInfo, principal principal.Principal, err error)
}

type OAuthProvider interface {
	// AuthenticateWithOAuth is oauth sign up/sign in.
	AuthenticateWithOAuth(oauthAuthInfo sso.AuthInfo, codeChallenge string, loginState sso.LoginState) (code *sso.SkygearAuthorizationCode, err error)

	// LinkOAuth links oauth identity with an existing user.
	LinkOAuth(oauthAuthInfo sso.AuthInfo, codeChallenge string, linkState sso.LinkState) (code *sso.SkygearAuthorizationCode, err error)

	ExtractAuthorizationCode(code *sso.SkygearAuthorizationCode) (authInfo *authinfo.AuthInfo, userProfile *userprofile.UserProfile, prin principal.Principal, err error)
}

// SessionProvider handles authentication process.
type SessionProvider interface {
	// BeginSession creates a new authentication session.
	BeginSession(client config.OAuthClientConfiguration, userID string, prin principal.Principal, reason session.CreateReason) (*Session, error)

	// StepSession update current step of an authentication session and return authentication result.
	StepSession(s *Session) (Result, error)

	// MakeResult loads related data for an existing session to create authentication result.
	MakeResult(client config.OAuthClientConfiguration, s *session.Session) (Result, error)

	// Resolve resolves token to authentication session.
	ResolveSession(token string) (*Session, error)
}
