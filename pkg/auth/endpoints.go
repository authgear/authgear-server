package auth

import (
	"net/url"
	"path"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
)

type EndpointsProvider struct {
	PrefixProvider urlprefix.Provider
}

func (p *EndpointsProvider) urlOf(relPath string) *url.URL {
	u := p.PrefixProvider.Value()
	u.Path = path.Join(u.Path, relPath)
	return u
}

func (p *EndpointsProvider) AuthorizeEndpointURI() *url.URL    { return p.urlOf("oauth2/authorize") }
func (p *EndpointsProvider) TokenEndpointURI() *url.URL        { return p.urlOf("oauth2/token") }
func (p *EndpointsProvider) RevokeEndpointURI() *url.URL       { return p.urlOf("oauth2/revoke") }
func (p *EndpointsProvider) JWKSEndpointURI() *url.URL         { return p.urlOf("oauth2/jwks") }
func (p *EndpointsProvider) UserInfoEndpointURI() *url.URL     { return p.urlOf("oauth2/userinfo") }
func (p *EndpointsProvider) EndSessionEndpointURI() *url.URL   { return p.urlOf("oauth2/end_session") }
func (p *EndpointsProvider) AuthenticateEndpointURI() *url.URL { return p.urlOf("./login") }
func (p *EndpointsProvider) PromoteUserEndpointURI() *url.URL  { return p.urlOf("./promote_user") }
func (p *EndpointsProvider) LogoutEndpointURI() *url.URL       { return p.urlOf("./logout") }
func (p *EndpointsProvider) SettingsEndpointURI() *url.URL     { return p.urlOf("./settings") }

var endpointsProviderSet = wire.NewSet(
	wire.Struct(new(EndpointsProvider), "*"),
)
