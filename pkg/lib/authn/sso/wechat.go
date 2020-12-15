package sso

import (
	"errors"
	"net/url"

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
	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("appid", w.ProviderConfig.ClientID)
	v.Add("redirect_uri", w.RedirectURL.SSOCallbackURL(w.ProviderConfig).String())
	v.Add("scope", w.ProviderConfig.Type.Scope())
	v.Add("state", param.State)

	authURL := wechatAuthorizationURL + "?" + v.Encode()
	v = url.Values{}
	v.Add("x_auth_url", authURL)
	return "/sso/wechat/auth?" + v.Encode(), nil
}

func (w *WechatImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (AuthInfo, error) {
	return AuthInfo{}, errors.New("not implemented")
}
