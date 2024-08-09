package config

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
		"id": { "type": "string" }
	},
	"required": ["id"]
}
`)

type SAMLConfig struct {
	Signing              *SAMLSigningConfig           `json:"signing,omitempty"`
	SAMLServiceProviders []*SAMLServiceProviderConfig `json:"service_providers,omitempty"`
}

func (c *SAMLConfig) ResolveProvider(id string) (*SAMLServiceProviderConfig, bool) {
	for _, sp := range c.SAMLServiceProviders {
		if sp.ID == id {
			return sp, true
		}
	}
	return nil, false
}

type SAMLServiceProviderConfig struct {
	ID string `json:"id,omitempty"`
}

var _ = Schema.Add("SAMLSigningConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"key_id": { "type": "string" },
		"signature_method": { "$ref": "#/$defs/SAMLSigningSignatureMethod" },
		"digest_method": { "$ref": "#/$defs/SAMLSigningDigestMethod" }
	}
}
`)

type SAMLSigningConfig struct {
	KeyID           string                     `json:"key_id,omitempty"`
	SignatureMethod SAMLSigningSignatureMethod `json:"signature_method,omitempty"`
	DigestMethod    SAMLSigningDigestMethod    `json:"digest_method,omitempty"`
}

func (c *SAMLSigningConfig) SetDefaults() {
	if c.SignatureMethod == "" {
		c.SignatureMethod = SAMLSigningSignatureMethodRSASHA256
	}
	if c.DigestMethod == "" {
		c.DigestMethod = SAMLSigningDigestMethodSHA256
	}
}

var _ = Schema.Add("SAMLSigningSignatureMethod", `
{
	"enum": ["RSAwithSHA256"] 
}
`)

type SAMLSigningSignatureMethod string

const (
	SAMLSigningSignatureMethodRSASHA256 SAMLSigningSignatureMethod = "RSAwithSHA256"
)

var _ = Schema.Add("SAMLSigningDigestMethod", `
{
	"enum": ["SHA256", "SHA1"] 
}
`)

type SAMLSigningDigestMethod string

const (
	SAMLSigningDigestMethodSHA256 SAMLSigningDigestMethod = "SHA256"
	SAMLSigningDigestMethodSHA1   SAMLSigningDigestMethod = "SHA1"
)
