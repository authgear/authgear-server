package samlbinding

import "net/url"

type SAMLBindingSigner interface {
	ConstructSignedQueryParameters(
		samlResponse string,
		relayState string,
	) (url.Values, error)
}
