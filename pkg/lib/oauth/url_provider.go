package oauth

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) FromWebAppURL(r protocol.AuthorizationRequest) *url.URL {
	return urlutil.WithQueryParamsAdded(p.Endpoints.FromWebAppEndpointURL(), r)
}
