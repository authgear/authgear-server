package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
)

type InputSelectIdentityOAuth interface {
	GetUserInfo() sso.AuthInfo
}

type EdgeSelectIdentityOAuth struct {
	Config config.OAuthSSOProviderConfig
}

type NodeSelectIdentityOAuth struct {
	UserInfo sso.AuthInfo `json:"auth_info"`
}
