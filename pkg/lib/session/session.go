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

type ListableSession interface {
	SessionID() string
	SessionType() Type

	GetClientID() string
	GetCreatedAt() time.Time
	GetAccessInfo() *access.Info
	GetDeviceInfo() (map[string]interface{}, bool)

	GetAuthenticationInfo() authenticationinfo.T

	ToAPIModel() *model.Session

	// SSOGroupIDPSessionID returns the IDP session id of the SSO group
	// if the session is not SSO enabled, SSOGroupIDPSessionID will be empty
	SSOGroupIDPSessionID() string
	// IsSameSSOGroup indicates whether the session is in the same SSO group
	IsSameSSOGroup(s ListableSession) bool
	Equal(s ListableSession) bool
}

type CreateReason string

const (
	CreateReasonSignup         CreateReason = "signup"
	CreateReasonLogin          CreateReason = "login"
	CreateReasonPromote        CreateReason = "promote"
	CreateReasonReauthenticate CreateReason = "reauthenticate"
)
