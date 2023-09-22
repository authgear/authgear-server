package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type mockRedirectURLProvider struct{}

var _ RedirectURLProvider = mockRedirectURLProvider{}

func (mockRedirectURLProvider) SSOCallbackURL(_ config.OAuthSSOProviderConfig) *url.URL {
	u, _ := url.Parse("https://localhost/")
	return u
}

type mockWechatURLProvider struct{}

var _ WechatURLProvider = mockWechatURLProvider{}

func (mockWechatURLProvider) WeChatAuthorizeURL(_ config.OAuthSSOProviderConfig) *url.URL {
	u, _ := url.Parse("https://localhost/wechat/authorize")
	return u
}

func (mockWechatURLProvider) WeChatCallbackEndpointURL() *url.URL {
	u, _ := url.Parse("https://localhost/wechat/callback")
	return u
}
