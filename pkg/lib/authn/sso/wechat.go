package sso

import (
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	wechatAuthorizationURL string = "https://open.weixin.qq.com/connect/oauth2/authorize"
)

type WechatURLProvider interface {
	AuthorizeEndpointURL(c config.OAuthSSOProviderConfig) *url.URL
	CallbackEndpointURL() *url.URL
}

type WechatImpl struct {
	ProviderConfig config.OAuthSSOProviderConfig
	Credentials    config.OAuthClientCredentialsItem
	URLProvider    WechatURLProvider
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
	v.Add("redirect_uri", w.URLProvider.CallbackEndpointURL().String())
	v.Add("scope", w.ProviderConfig.Type.Scope())
	v.Add("state", param.State)
	prompt := w.GetPrompt(param.Prompt)
	if len(prompt) > 0 {
		v.Add("prompt", strings.Join(prompt, " "))
	}

	authURL := wechatAuthorizationURL + "?" + v.Encode()
	v = url.Values{}
	v.Add("x_auth_url", authURL)
	return w.URLProvider.AuthorizeEndpointURL(w.ProviderConfig).String() + "?" + v.Encode(), nil
}

func (w *WechatImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (AuthInfo, error) {
	return w.NonOpenIDConnectGetAuthInfo(r, param)
}

func (w *WechatImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, _ GetAuthInfoParam) (authInfo AuthInfo, err error) {
	accessTokenResp, err := wechatFetchAccessTokenResp(
		r.Code,
		w.ProviderConfig.ClientID,
		w.Credentials.ClientSecret,
	)
	if err != nil {
		return
	}

	rawProfile, err := wechatFetchUserProfile(accessTokenResp)
	if err != nil {
		return
	}

	config := w.Config()
	var userID string
	if config.IsSandboxAccount {
		if accessTokenResp.UnionID() != "" {
			err = InvalidConfiguration.New("invalid is_sandbox_account config, WeChat sandbox account should not have union id")
			return
		}
		userID = accessTokenResp.OpenID()
	} else {
		userID = accessTokenResp.UnionID()
	}

	if userID == "" {
		// this may happen if developer misconfigure is_sandbox_account, e.g. sandbox account doesn't have union id
		err = InvalidConfiguration.New("invalid is_sandbox_account config, missing user id in wechat token response")
		return
	}

	authInfo.ProviderConfig = w.ProviderConfig
	authInfo.ProviderAccessTokenResp = accessTokenResp
	authInfo.ProviderRawProfile = rawProfile
	authInfo.ProviderUserInfo = ProviderUserInfo{
		ID: userID,
	}
	return
}

func (w *WechatImpl) GetPrompt(prompt []string) []string {
	return []string{}
}

var (
	_ OAuthProvider            = &WechatImpl{}
	_ NonOpenIDConnectProvider = &WechatImpl{}
)
