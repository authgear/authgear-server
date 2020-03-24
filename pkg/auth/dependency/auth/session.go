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
	GetAccessInfo() *AccessInfo
	ToAPIModel() *model.Session
}
