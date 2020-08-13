package auth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/lib/api/model"
)

const (
	SessionTypeIdentityProvider authn.SessionType = "idp"
	SessionTypeOfflineGrant     authn.SessionType = "offline_grant"
)

// nolint: golint
type AuthSession interface {
	authn.Session
	GetClientID() string
	GetCreatedAt() time.Time
	GetAccessInfo() *AccessInfo
	ToAPIModel() *model.Session
}

type SessionDeleteReason string

const (
	SessionDeleteReasonLogout SessionDeleteReason = "logout"
	SessionDeleteReasonRevoke SessionDeleteReason = "revoke"
)

type SessionCreateReason string

const (
	SessionCreateReasonSignup  SessionCreateReason = "signup"
	SessionCreateReasonLogin   SessionCreateReason = "login"
	SessionCreateReasonPromote SessionCreateReason = "promote"
)
