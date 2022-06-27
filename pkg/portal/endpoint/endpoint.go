package endpoint

import (
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type OriginProvider interface {
	Origin() *url.URL
}

type RequestOriginProvider struct {
	HTTPHost  httputil.HTTPHost
	HTTPProto httputil.HTTPProto
}

func (p *RequestOriginProvider) Origin() *url.URL {
	return &url.URL{
		Host:   string(p.HTTPHost),
		Scheme: string(p.HTTPProto),
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

func (p *EndpointsProvider) AcceptCollaboratorInvitationEndpointURL() *url.URL {
	return p.urlOf("collaborators/invitation")
}

func (p *EndpointsProvider) BillingEndpointURL(relayGlobalAppID string) *url.URL {
	return p.urlOf(path.Join("project", relayGlobalAppID, "billing"))
}

func (p *EndpointsProvider) BillingRedirectEndpointURL(relayGlobalAppID string) *url.URL {
	return p.urlOf(path.Join("project", relayGlobalAppID, "billing-redirect"))
}
