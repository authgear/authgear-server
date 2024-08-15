package samlsession

import (
	"net/url"
)

const (
	queryNameSAMLSessionID string = "x_saml_session_id"
)

type UIURLBuilderAuthUIEndpointsProvider interface {
	OAuthEntrypointURL() *url.URL
}

type UIURLBuilder struct {
	Endpoints UIURLBuilderAuthUIEndpointsProvider
}

func (b *UIURLBuilder) BuildAuthenticationURL(s *SAMLSession) (*url.URL, error) {
	endpoint := b.Endpoints.OAuthEntrypointURL()

	q := endpoint.Query()
	q.Set(queryNameSAMLSessionID, s.ID)
	endpoint.RawQuery = q.Encode()
	return endpoint, nil
}
