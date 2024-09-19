package samlbinding

import "net/url"

type SAMLRedirectBindingSigner interface {
	ConstructSignedQueryParameters(
		samlResponse string,
		relayState string,
	) (url.Values, error)
}
