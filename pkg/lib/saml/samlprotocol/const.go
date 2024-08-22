package samlprotocol

import (
	"fmt"

	crewjamsaml "github.com/crewjam/saml"
)

type SAMLBinding string

const (
	SAMLBindingHTTPRedirect SAMLBinding = crewjamsaml.HTTPRedirectBinding
	SAMLBindingHTTPPost     SAMLBinding = crewjamsaml.HTTPPostBinding
)

var SupportedBindings []SAMLBinding = []SAMLBinding{
	SAMLBindingHTTPRedirect,
	SAMLBindingHTTPPost,
}

func (b SAMLBinding) IsSupported() bool {
	for _, supported := range SupportedBindings {
		if b == supported {
			return true
		}
	}
	return false
}

const xmlSchemaNamespace = "xs"

const (
	// https://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf 3.2.2
	SAMLVersion2 string = "2.0"
)

const (
	SAMLIssertFormatEntity = "urn:oasis:names:tc:SAML:2.0:nameid-format:entity"
)

const (
	SAMLAttrnameFormatBasic = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
)

var (
	SAMLAttrTypeString = fmt.Sprintf("%s:string", xmlSchemaNamespace)
)

const timeFormat = "2006-01-02T15:04:05.999Z07:00"
const canonicalizerPrefixList = ""
