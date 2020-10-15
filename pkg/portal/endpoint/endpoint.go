package endpoint

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

type RequestOriginProvider struct {
	Request    *http.Request
	TrustProxy config.TrustProxy
}

func (p *RequestOriginProvider) Origin() *url.URL {
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

func (p *EndpointsProvider) AcceptCollaboratorInvitationEndpointURL() *url.URL {
	return p.urlOf("collaborators/invitation")
}
