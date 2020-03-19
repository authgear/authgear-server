package auth

import (
	"net/url"
	"path"

	"github.com/google/wire"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
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
func (p *EndpointsProvider) AuthenticateEndpointURI() *url.URL { return p.urlOf(".") }

var endpointsProviderSet = wire.NewSet(
	wire.Struct(new(EndpointsProvider), "*"),
	wire.Bind(new(oauthhandler.AuthorizeEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oauthhandler.AuthenticateEndpointProvider), new(*EndpointsProvider)),
)
