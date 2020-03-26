package oidc

import "net/url"

type JWKSEndpointProvider interface {
	JWKSEndpointURI() *url.URL
}

type UserInfoEndpointProvider interface {
	UserInfoEndpointURI() *url.URL
}

type EndSessionEndpointProvider interface {
	EndSessionEndpointURI() *url.URL
}
