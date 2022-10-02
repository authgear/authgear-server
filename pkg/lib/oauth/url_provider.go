package oauth

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) ConsentURL(r protocol.AuthorizationRequest) *url.URL {
	return p.Endpoints.ConsentEndpointURL()
}
