package oauth

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) AuthorizeURL(r protocol.AuthorizationRequest) *url.URL {
	return urlutil.WithQueryParamsAdded(p.Endpoints.AuthorizeEndpointURL(), r)
}
