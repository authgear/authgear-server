package binding

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
