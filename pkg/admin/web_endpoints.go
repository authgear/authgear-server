package admin

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type BaseURLProvider struct {
	HTTP *config.HTTPConfig
}

func (p *BaseURLProvider) BaseURL() *url.URL {
	u, err := url.Parse(p.HTTP.PublicOrigin)
	if err != nil {
		panic(err)
	}
	return u
}

func (*BaseURLProvider) MagicLinkVerificationEndpointURL() *url.URL {
	panic("not implemented")
}

type OAuthURLProvider struct{}

func (*OAuthURLProvider) AuthorizeEndpointURL() *url.URL {
	panic("not implemented")
}

func (*OAuthURLProvider) ConsentEndpointURL() *url.URL {
	panic("not implemented")
}

func (*OAuthURLProvider) TokenEndpointURL() *url.URL {
	panic("not implemented")
}

func (*OAuthURLProvider) RevokeEndpointURL() *url.URL {
	panic("not implemented")
}

type WechatURLProvider struct{}

func (*WechatURLProvider) WeChatAuthorizeURL(config.OAuthSSOProviderConfig) *url.URL {
	panic("not implemented")
}

func (*WechatURLProvider) WeChatCallbackEndpointURL() *url.URL {
	panic("not implemented")
}

type ResetPasswordURLProvider struct{}

func (*ResetPasswordURLProvider) ResetPasswordURL(code string) *url.URL {
	panic("not implemented")
}

type SSOCallbackURLProvider struct{}

func (*SSOCallbackURLProvider) SSOCallbackURL(providerConfig config.OAuthSSOProviderConfig) *url.URL {
	panic("not implemented")
}
