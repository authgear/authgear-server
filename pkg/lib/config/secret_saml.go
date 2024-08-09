package config

var _ = SecretConfigSchema.Add("SAMLIdpSigningMaterials", `
{
	"type": "object",
	"properties": {
		"certs": {
			"type": "array",
			"items": { "$ref": "#/$defs/SAMLIdpSigningCert" },
			"minItems": 1,
			"maxItems": 2
		}
	},
	"required": ["certs"]
}
`)

type SAMLIdpSigningMaterials struct {
	Certs []*SAMLIdpSigningCert `json:"certs,omitempty"`
}

var _ SecretItemData = &SAMLIdpSigningMaterials{}

func (s *SAMLIdpSigningMaterials) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("SAMLIdpSigningCert", `
{
	"type": "object",
	"properties": {
		"cert": { "$ref": "#/$defs/X509Cert" },
		"key": { "$ref": "#/$defs/JWK" }
	},
	"required": ["cert", "key"]
}
`)

type SAMLIdpSigningCert struct {
	Cert *X509Cert `json:"cert,omitempty"`
	Key  *JWK      `json:"key,omitempty"`
}

var _ = SecretConfigSchema.Add("SAMLSpSigningMaterials", `
{
	"type": "array",
	"items": { "$ref": "#/$defs/SAMLSpSigningCert" }
}
`)

type SAMLSpSigningMaterials []SAMLSpSigningCert

var _ SecretItemData = &SAMLSpSigningMaterials{}

func (s *SAMLSpSigningMaterials) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("SAMLSpSigningCert", `
{
	"type": "object",
	"properties": {
		"service_provider_id": { "type": "string" },
		"certs": {
			"type": "array",
			"items": { "$ref": "#/$defs/X509Cert" }
		}
	},
	"required": ["service_provider_id", "certs"]
}
`)

type SAMLSpSigningCert struct {
	ServiceProviderID string     `json:"service_provider_id,omitempty"`
	Certs             []X509Cert `json:"certs,omitempty"`
}
