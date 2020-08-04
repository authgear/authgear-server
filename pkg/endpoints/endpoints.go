package endpoints

import (
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type Provider struct {
	Request *http.Request
	Config  *config.ServerConfig
}

func (p *Provider) BaseURL() *url.URL {
	return &url.URL{
		Host:   httputil.GetHost(p.Request, p.Config.TrustProxy),
		Scheme: httputil.GetProto(p.Request, p.Config.TrustProxy),
	}
}

func (p *Provider) urlOf(relPath string) *url.URL {
	u := p.BaseURL()
	u.Path = path.Join(u.Path, relPath)
	return u
}

func (p *Provider) AuthorizeEndpointURL() *url.URL     { return p.urlOf("oauth2/authorize") }
func (p *Provider) TokenEndpointURL() *url.URL         { return p.urlOf("oauth2/token") }
func (p *Provider) RevokeEndpointURL() *url.URL        { return p.urlOf("oauth2/revoke") }
func (p *Provider) JWKSEndpointURL() *url.URL          { return p.urlOf("oauth2/jwks") }
func (p *Provider) UserInfoEndpointURL() *url.URL      { return p.urlOf("oauth2/userinfo") }
func (p *Provider) EndSessionEndpointURL() *url.URL    { return p.urlOf("oauth2/end_session") }
func (p *Provider) AuthenticateEndpointURL() *url.URL  { return p.urlOf("./login") }
func (p *Provider) PromoteUserEndpointURL() *url.URL   { return p.urlOf("./promote_user") }
func (p *Provider) LogoutEndpointURL() *url.URL        { return p.urlOf("./logout") }
func (p *Provider) SettingsEndpointURL() *url.URL      { return p.urlOf("./settings") }
func (p *Provider) ResetPasswordEndpointURL() *url.URL { return p.urlOf("./reset_password") }
func (p *Provider) VerifyUserEndpointURL() *url.URL    { return p.urlOf("./verify_user") }
func (p *Provider) SSOCallbackEndpointURL() *url.URL   { return p.urlOf("sso/oauth2/callback") }
