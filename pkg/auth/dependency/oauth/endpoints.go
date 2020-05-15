package oauth

import "net/url"

type AuthorizeEndpointProvider interface {
	AuthorizeEndpointURI() *url.URL
}

type TokenEndpointProvider interface {
	TokenEndpointURI() *url.URL
}

type RevokeEndpointProvider interface {
	RevokeEndpointURI() *url.URL
}
