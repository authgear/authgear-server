package config

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	dsig "github.com/russellhaering/goxmldsig"
)

var _ = Schema.Add("SAMLConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"signing": { "$ref": "#/$defs/SAMLSigningConfig" },
		"service_providers": {
			"type": "array",
			"items": { "$ref": "#/$defs/SAMLServiceProviderConfig" }
		}
	}
}
`)

var _ = Schema.Add("SAMLServiceProviderConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"id": { "type": "string" },
		"nameid_format": { "$ref": "#/$defs/SAMLNameIDFormat" },
		"nameid_attribute_pointer": { "$ref": "#/$defs/SAMLNameIDAttributePointer" },
		"acs_urls": {
			"type": "array",
			"items": { "type": "string", "format": "uri" },
			"minItems": 1
		}
	},
	"required": ["id", "acs_urls"]
}
`)

type SAMLConfig struct {
	Signing          *SAMLSigningConfig           `json:"signing,omitempty"`
	ServiceProviders []*SAMLServiceProviderConfig `json:"service_providers,omitempty"`
}

func (c *SAMLConfig) ResolveProvider(id string) (*SAMLServiceProviderConfig, bool) {
	for _, sp := range c.ServiceProviders {
		if sp.ID == id {
			return sp, true
		}
	}
	return nil, false
}

var _ = Schema.Add("SAMLNameIDFormat", `
{
	"type": "string",
	"enum": [
		"urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
		"urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
	]
}
`)

type SAMLNameIDFormat string

const (
	SAMLNameIDFormatUnspecified  SAMLNameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
	SAMLNameIDFormatEmailAddress SAMLNameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
)

var _ = Schema.Add("SAMLNameIDAttributePointer", `
{
	"type": "string",
	"format": "json-pointer",
	"enum": [
		"/sub",
		"/email",
		"/phone_number",
		"/preferred_username"
	]
}
`)

type SAMLNameIDAttributePointer string

func (p SAMLNameIDAttributePointer) MustGetJSONPointer() jsonpointer.T {
	pointer := jsonpointer.MustParse(string(p))
	return pointer
}

type SAMLServiceProviderConfig struct {
	ID                     string                     `json:"id,omitempty"`
	NameIDFormat           SAMLNameIDFormat           `json:"nameid_format,omitempty"`
	NameIDAttributePointer SAMLNameIDAttributePointer `json:"nameid_attribute_pointer,omitempty"`
	AcsURLs                []string                   `json:"acs_urls,omitempty"`
}

func (c *SAMLServiceProviderConfig) SetDefaults() {
	if c.NameIDFormat == "" {
		c.NameIDFormat = SAMLNameIDFormatUnspecified
	}

	if c.NameIDFormat == SAMLNameIDFormatUnspecified && c.NameIDAttributePointer == "" {
		c.NameIDAttributePointer = "/sub"
	}
}

func (c *SAMLServiceProviderConfig) DefaultAcsURL() string {
	return c.AcsURLs[0]
}

var _ = Schema.Add("SAMLSigningConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"key_id": { "type": "string" },
		"signature_algorithm": { "$ref": "#/$defs/SAMLSigningSignatureAlgorithm" }
	}
}
`)

type SAMLSigningConfig struct {
	KeyID              string                        `json:"key_id,omitempty"`
	SignatureAlgorithm SAMLSigningSignatureAlgorithm `json:"signature_algorithm,omitempty"`
}

func (c *SAMLSigningConfig) SetDefaults() {
	if c.SignatureAlgorithm == "" {
		c.SignatureAlgorithm = SAMLSigningSignatureMethodRSASHA256
	}
}

var _ = Schema.Add("SAMLSigningSignatureAlgorithm", `
{
	"enum": ["http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"] 
}
`)

type SAMLSigningSignatureAlgorithm string

func (m SAMLSigningSignatureAlgorithm) ToDsigSignatureMethod() string {
	switch m {
	case SAMLSigningSignatureMethodRSASHA256:
		return dsig.RSASHA256SignatureMethod
	default:
		panic("unknown signature method")
	}
}

const (
	SAMLSigningSignatureMethodRSASHA256 SAMLSigningSignatureAlgorithm = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
)

var _ = Schema.Add("SAMLSigningDigestMethod", `
{
	"enum": ["SHA256"] 
}
`)

type SAMLSigningDigestMethod string

const (
	SAMLSigningDigestMethodSHA256 SAMLSigningDigestMethod = "SHA256"
)
