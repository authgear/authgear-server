package sso

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	wechatAuthorizationURL string = "https://open.weixin.qq.com/connect/oauth2/authorize"
)

type WechatImpl struct {
	RedirectURL    RedirectURLProvider
	ProviderConfig config.OAuthSSOProviderConfig
	Credentials    config.OAuthClientCredentialsItem
}

func (*WechatImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeWechat
}

func (w *WechatImpl) Config() config.OAuthSSOProviderConfig {
	return w.ProviderConfig
}

func (w *WechatImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	return "", errors.New("not implemented")
}

func (w *WechatImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (AuthInfo, error) {
	return AuthInfo{}, errors.New("not implemented")
}
