package oidc

import "net/url"

type EndpointsProvider interface {
	BaseURL() *url.URL
	JWKSEndpointURL() *url.URL
	UserInfoEndpointURL() *url.URL
	EndSessionEndpointURL() *url.URL
}
