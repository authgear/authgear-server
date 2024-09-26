package samlbinding

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/saml"
)

type SAMLRedirectBindingSigner interface {
	ConstructSignedQueryParameters(
		relayState string,
		el *saml.SAMLElementToSign,
	) (url.Values, error)
}
