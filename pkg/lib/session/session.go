package session

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type Type string

const (
	TypeIdentityProvider Type = "idp"
	TypeOfflineGrant     Type = "offline_grant"
)

// TokenType records the credential a ResolvedSession was resolved from --
// distinct from Type, which records the underlying session kind (idp vs
// offline_grant). It lets /resolve decide, without a side-channel context
// value, whether the presented credential is subject to the third-party
// opaque-access-token gate (see pkg/resolver/handler/resolve.go):
// TokenTypeJWT and TokenTypeOpaque are the two possible shapes of a bearer
// access token (Authorization header / app access token cookie), while
// TokenTypeCookies (IDP session cookie) and TokenTypeAppSession (app
// session token cookie) are both always-accept, kept distinct only for
// observability.
type TokenType string

const (
	TokenTypeCookies    TokenType = "cookies"
	TokenTypeJWT        TokenType = "jwt"
	TokenTypeOpaque     TokenType = "opaque"
	TokenTypeAppSession TokenType = "app_session"
)

type SessionBase interface {
	SessionID() string
	SessionType() Type
	GetAuthenticationInfo() authenticationinfo.T
	// SSOGroupIDPSessionID returns the IDP session id of the SSO group
	// if the session is not SSO enabled, SSOGroupIDPSessionID will be empty
	SSOGroupIDPSessionID() string
}

type ResolvedSession interface {
	SessionBase
	Session()
	GetCreatedAt() time.Time
	GetExpireAt() time.Time
	GetAccessInfo() *access.Info
	CreateNewAuthenticationInfoByThisSession() authenticationinfo.T
}

type ListableSession interface {
	SessionBase
	ListableSession()
	GetCreatedAt() time.Time
	GetAccessInfo() *access.Info
	GetDeviceInfo() (map[string]any, bool)

	ToAPIModel() *model.Session

	// IsSameSSOGroup indicates whether the session is in the same SSO group
	IsSameSSOGroup(s SessionBase) bool
	EqualSession(s SessionBase) bool

	GetParticipatedSAMLServiceProviderIDsSet() setutil.Set[string]
}

type CreateReason string

const (
	CreateReasonSignup         CreateReason = "signup"
	CreateReasonLogin          CreateReason = "login"
	CreateReasonPromote        CreateReason = "promote"
	CreateReasonReauthenticate CreateReason = "reauthenticate"
)
