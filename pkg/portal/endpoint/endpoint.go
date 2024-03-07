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

func (p *EndpointsProvider) urlOf(relPath string) *url.URL {
	// If we do not set Path = "/", then in urlOf,
	// Path will have no leading /.
	// It is problematic when Path is used in comparison.
	//
	// u, _ := url.Parse("https://example.com/path")
	// // u.Path is "/path"
	// uu := endpoints.urlOf("path")
	// // uu.Path is "path"
	// So direct comparison will yield a surprising result.
	// More confusing is that u.String() == uu.String()
	// Because String() will add leading / to make the URL legal.
	u := p.OriginProvider.Origin()
	u.Path = path.Join("/", relPath)
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
