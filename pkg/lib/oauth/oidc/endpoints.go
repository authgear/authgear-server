package oidc

import "net/url"

type EndpointsProvider interface {
	Origin() *url.URL
	JWKSEndpointURL() *url.URL
	UserInfoEndpointURL() *url.URL
	EndSessionEndpointURL() *url.URL
}
