package samlprotocol

import (
	crewjamsaml "github.com/crewjam/saml"
)

type SAMLBinding string

const (
	SAMLBindingHTTPRedirect SAMLBinding = crewjamsaml.HTTPRedirectBinding
	SAMLBindingPostRedirect SAMLBinding = crewjamsaml.HTTPPostBinding
)

var SupportedBindings []SAMLBinding = []SAMLBinding{
	SAMLBindingHTTPRedirect,
	SAMLBindingPostRedirect,
}

func (b SAMLBinding) IsSupported() bool {
	for _, supported := range SupportedBindings {
		if b == supported {
			return true
		}
	}
	return false
}

const (
	// https://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf 3.2.2
	SAMLVersion2 string = "2.0"
)
