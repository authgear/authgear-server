package session

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

type Type string

const (
	TypeIdentityProvider Type = "idp"
	TypeOfflineGrant     Type = "offline_grant"
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
	GetDeviceInfo() (map[string]interface{}, bool)

	ToAPIModel() *model.Session

	// IsSameSSOGroup indicates whether the session is in the same SSO group
	IsSameSSOGroup(s SessionBase) bool
	EqualSession(s SessionBase) bool
}

type CreateReason string

const (
	CreateReasonSignup         CreateReason = "signup"
	CreateReasonLogin          CreateReason = "login"
	CreateReasonPromote        CreateReason = "promote"
	CreateReasonReauthenticate CreateReason = "reauthenticate"
)
