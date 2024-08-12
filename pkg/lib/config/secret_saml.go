package config

var _ = SecretConfigSchema.Add("SAMLIdpSigningMaterials", `
{
	"type": "object",
	"properties": {
		"certificates": {
			"type": "array",
			"items": { "$ref": "#/$defs/SAMLIdpSigningCert" },
			"minItems": 1,
			"maxItems": 2
		}
	},
	"required": ["certificates"]
}
`)

type SAMLIdpSigningMaterials struct {
	Certificates []*SAMLIdpSigningCert `json:"certificates,omitempty"`
}

func (m *SAMLIdpSigningMaterials) FindSigningCert(keyID string) (*SAMLIdpSigningCert, bool) {
	for _, cert := range m.Certificates {
		if cert.Key.KeyID() == keyID {
			return cert, true
		}
	}
	return nil, false
}

var _ SecretItemData = &SAMLIdpSigningMaterials{}

func (s *SAMLIdpSigningMaterials) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("SAMLIdpSigningCert", `
{
	"type": "object",
	"properties": {
		"certificate": { "$ref": "#/$defs/X509Cert" },
		"key": { "$ref": "#/$defs/JWK" }
	},
	"required": ["certificate", "key"]
}
`)

type SAMLIdpSigningCert struct {
	Certificate *X509Cert `json:"certificate,omitempty"`
	Key         *JWK      `json:"key,omitempty"`
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
		"certificates": {
			"type": "array",
			"items": { "$ref": "#/$defs/X509Cert" }
		}
	},
	"required": ["service_provider_id", "certificates"]
}
`)

type SAMLSpSigningCert struct {
	ServiceProviderID string     `json:"service_provider_id,omitempty"`
	Certificates      []X509Cert `json:"certificates,omitempty"`
}
