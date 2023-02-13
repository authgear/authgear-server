package webapp

import (
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type WechatEndpoints interface {
	WeChatAuthorizeEndpointURL() *url.URL
	WeChatCallbackEndpointURL() *url.URL
}

type WechatURLProvider struct {
	Endpoints WechatEndpoints
}

func (p *WechatURLProvider) AuthorizeEndpointURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := p.Endpoints.WeChatAuthorizeEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}

func (p *WechatURLProvider) CallbackEndpointURL() *url.URL {
	return p.Endpoints.WeChatCallbackEndpointURL()
}
