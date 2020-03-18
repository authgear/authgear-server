package oauth

import "net/url"

type URIProvider interface {
	AuthorizeURI() *url.URL
	AuthenticateURI() *url.URL
}
