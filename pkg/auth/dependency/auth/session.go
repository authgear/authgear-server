package auth

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

const (
	SessionTypeIdentityProvider authn.SessionType = "idp"
	SessionTypeOfflineGrant     authn.SessionType = "offline_grant"
)

// nolint: golint
type AuthSession interface {
	authn.Session
	GetClientID() string
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
	SessionCreateReasonSignup SessionCreateReason = "signup"
	SessionCreateReasonLogin  SessionCreateReason = "login"
)
