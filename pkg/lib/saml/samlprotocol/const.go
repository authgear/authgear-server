package samlprotocol

import (
	"fmt"
)

type SAMLBinding string

const (
	SAMLBindingHTTPRedirect SAMLBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
	SAMLBindingHTTPPost     SAMLBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
)

var SSOSupportedBindings []SAMLBinding = []SAMLBinding{
	SAMLBindingHTTPRedirect,
	SAMLBindingHTTPPost,
}

var SLOSupportedBindings []SAMLBinding = []SAMLBinding{
	SAMLBindingHTTPRedirect,
	SAMLBindingHTTPPost,
}

var ACSSupportedBindings []SAMLBinding = []SAMLBinding{
	SAMLBindingHTTPPost,
}

func (b SAMLBinding) IsACSSupported() bool {
	for _, supported := range ACSSupportedBindings {
		if b == supported {
			return true
		}
	}
	return false
}

type SAMLNameIDFormat string

const (
	SAMLNameIDFormatUnspecified  SAMLNameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
	SAMLNameIDFormatEmailAddress SAMLNameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
)

const xmlSchemaNamespace = "xs"
const dsigSchemaNamespace = "ds"

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
