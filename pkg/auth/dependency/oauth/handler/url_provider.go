package handler

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

type EndpointsProvider interface {
	AuthorizeEndpointURI() *url.URL
}

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) AuthorizeURI(r protocol.AuthorizationRequest) *url.URL {
	return coreurl.WithQueryParamsAdded(p.Endpoints.AuthorizeEndpointURI(), r)
}
