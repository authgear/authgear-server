package auth

import (
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type OriginProvider interface {
	Origin() *url.URL
}

type MainOriginProvider struct {
	Request    *http.Request
	TrustProxy config.TrustProxy
}

func (p *MainOriginProvider) Origin() *url.URL {
	return &url.URL{
		Host:   httputil.GetHost(p.Request, bool(p.TrustProxy)),
		Scheme: httputil.GetProto(p.Request, bool(p.TrustProxy)),
	}
}

type EndpointsProvider struct {
	OriginProvider OriginProvider
}

func (p *EndpointsProvider) BaseURL() *url.URL {
	return p.OriginProvider.Origin()
}

func (p *EndpointsProvider) urlOf(relPath string) *url.URL {
	u := p.BaseURL()
	u.Path = path.Join(u.Path, relPath)
	return u
}

func (p *EndpointsProvider) AuthorizeEndpointURL() *url.URL      { return p.urlOf("oauth2/authorize") }
func (p *EndpointsProvider) TokenEndpointURL() *url.URL          { return p.urlOf("oauth2/token") }
func (p *EndpointsProvider) RevokeEndpointURL() *url.URL         { return p.urlOf("oauth2/revoke") }
func (p *EndpointsProvider) JWKSEndpointURL() *url.URL           { return p.urlOf("oauth2/jwks") }
func (p *EndpointsProvider) UserInfoEndpointURL() *url.URL       { return p.urlOf("oauth2/userinfo") }
func (p *EndpointsProvider) EndSessionEndpointURL() *url.URL     { return p.urlOf("oauth2/end_session") }
func (p *EndpointsProvider) LoginEndpointURL() *url.URL          { return p.urlOf("./login") }
func (p *EndpointsProvider) SignupEndpointURL() *url.URL         { return p.urlOf("./signup") }
func (p *EndpointsProvider) PromoteUserEndpointURL() *url.URL    { return p.urlOf("./promote_user") }
func (p *EndpointsProvider) LogoutEndpointURL() *url.URL         { return p.urlOf("./logout") }
func (p *EndpointsProvider) SettingsEndpointURL() *url.URL       { return p.urlOf("./settings") }
func (p *EndpointsProvider) ResetPasswordEndpointURL() *url.URL  { return p.urlOf("./reset_password") }
func (p *EndpointsProvider) VerifyIdentityEndpointURL() *url.URL { return p.urlOf("./verify_identity") }
func (p *EndpointsProvider) SSOCallbackEndpointURL() *url.URL    { return p.urlOf("sso/oauth2/callback") }

func (p *EndpointsProvider) WeChatAuthorizeEndpointURL() *url.URL { return p.urlOf("sso/wechat/auth") }
func (p *EndpointsProvider) WeChatCallbackEndpointURL() *url.URL {
	return p.urlOf("sso/wechat/callback")
}
