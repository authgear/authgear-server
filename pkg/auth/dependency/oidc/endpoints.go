package oidc

import "net/url"

type JWKSEndpointProvider interface {
	JWKSEndpointURI() *url.URL
}

type UserInfoEndpointProvider interface {
	TokenEndpointURI() *url.URL
}
