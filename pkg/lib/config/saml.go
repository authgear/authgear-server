package config

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
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
		"client_id": { "type": "string" },
		"nameid_format": { "$ref": "#/$defs/SAMLNameIDFormat" },
		"nameid_attribute_pointer": { "$ref": "#/$defs/SAMLNameIDAttributePointer" },
		"acs_urls": {
			"type": "array",
			"items": { "type": "string", "format": "uri" },
			"minItems": 1
		},
		"destination": { "type": "string", "format": "uri" },
		"recipient": { "type": "string", "format": "uri" },
		"audience": { "type": "string", "format": "uri" },
		"assertion_valid_duration":  { "$ref": "#/$defs/DurationString" }
	},
	"required": ["acs_urls"]
}
`)

type SAMLConfig struct {
	Signing          *SAMLSigningConfig           `json:"signing,omitempty"`
	ServiceProviders []*SAMLServiceProviderConfig `json:"service_providers,omitempty"`
}

func (c *SAMLConfig) ResolveProvider(id string) (*SAMLServiceProviderConfig, bool) {
	if id == "" {
		return nil, false
	}
	for _, sp := range c.ServiceProviders {
		if sp.ID == id || sp.ClientID == id {
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
	ClientID               string                     `json:"client_id,omitempty"`
	NameIDFormat           SAMLNameIDFormat           `json:"nameid_format,omitempty"`
	NameIDAttributePointer SAMLNameIDAttributePointer `json:"nameid_attribute_pointer,omitempty"`
	AcsURLs                []string                   `json:"acs_urls,omitempty"`
	Destination            string                     `json:"destination,omitempty"`
	Recipient              string                     `json:"recipient,omitempty"`
	Audience               string                     `json:"audience,omitempty"`
	AssertionValidDuration DurationString             `json:"assertion_valid_duration,omitempty"`
}

func (c *SAMLServiceProviderConfig) SetDefaults() {
	if c.NameIDFormat == "" {
		c.NameIDFormat = SAMLNameIDFormatUnspecified
	}

	if c.NameIDFormat == SAMLNameIDFormatUnspecified && c.NameIDAttributePointer == "" {
		c.NameIDAttributePointer = "/sub"
	}

	if c.AssertionValidDuration == "" {
		c.AssertionValidDuration = DurationString("20m")
	}
}

func (c *SAMLServiceProviderConfig) DefaultAcsURL() string {
	return c.AcsURLs[0]
}

func (c *SAMLServiceProviderConfig) GetID() string {
	if c.ID != "" {
		return c.ID
	}
	if c.ClientID != "" {
		return c.ClientID
	}
	panic("unexpected: service provider does not have id nor client id")
}

var _ = Schema.Add("SAMLSigningConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"key_id": { "type": "string" },
		"signature_method": { "$ref": "#/$defs/SAMLSigningSignatureMethod" }
	}
}
`)

type SAMLSigningConfig struct {
	KeyID           string                     `json:"key_id,omitempty"`
	SignatureMethod SAMLSigningSignatureMethod `json:"signature_method,omitempty"`
}

func (c *SAMLSigningConfig) SetDefaults() {
	if c.SignatureMethod == "" {
		c.SignatureMethod = SAMLSigningSignatureMethodRSASHA256
	}
}

var _ = Schema.Add("SAMLSigningSignatureMethod", `
{
	"enum": ["http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"] 
}
`)

type SAMLSigningSignatureMethod string

const (
	SAMLSigningSignatureMethodRSASHA256 SAMLSigningSignatureMethod = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
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
