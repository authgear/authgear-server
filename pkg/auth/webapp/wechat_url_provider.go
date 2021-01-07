package webapp

import (
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type WechatURLProvider struct {
	Endpoints EndpointsProvider
}

func (p *WechatURLProvider) AuthorizeEndpointURL() *url.URL {
	return p.Endpoints.WeChatAuthorizeEndpointURL()
}

func (p *WechatURLProvider) SSOCallbackURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := p.Endpoints.WeChatCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}
