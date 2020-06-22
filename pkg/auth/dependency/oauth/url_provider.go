package oauth

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) AuthorizeURL(r protocol.AuthorizationRequest) *url.URL {
	return coreurl.WithQueryParamsAdded(p.Endpoints.AuthorizeEndpointURL(), r)
}
