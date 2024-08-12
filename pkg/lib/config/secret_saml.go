package config

var _ = SecretConfigSchema.Add("SAMLIdpSigningMaterials", `
{
	"type": "object",
	"properties": {
		"certificates": {
			"type": "array",
			"items": { "$ref": "#/$defs/SAMLIdpSigningCertificate" },
			"minItems": 1,
			"maxItems": 2
		}
	},
	"required": ["certificates"]
}
`)

type SAMLIdpSigningMaterials struct {
	Certificates []*SAMLIdpSigningCertificate `json:"certificates,omitempty"`
}

func (m *SAMLIdpSigningMaterials) FindSigningCert(keyID string) (*SAMLIdpSigningCertificate, bool) {
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

var _ = SecretConfigSchema.Add("SAMLIdpSigningCertificate", `
{
	"type": "object",
	"properties": {
		"certificate": { "$ref": "#/$defs/X509Certtificate" },
		"key": { "$ref": "#/$defs/JWK" }
	},
	"required": ["certificate", "key"]
}
`)

type SAMLIdpSigningCertificate struct {
	Certificate *X509Certtificate `json:"certificate,omitempty"`
	Key         *JWK              `json:"key,omitempty"`
}

var _ = SecretConfigSchema.Add("SAMLSpSigningMaterials", `
{
	"type": "array",
	"items": { "$ref": "#/$defs/SAMLSpSigningCertificate" }
}
`)

type SAMLSpSigningMaterials []SAMLSpSigningCertificate

var _ SecretItemData = &SAMLSpSigningMaterials{}

func (s *SAMLSpSigningMaterials) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("SAMLSpSigningCertificate", `
{
	"type": "object",
	"properties": {
		"service_provider_id": { "type": "string" },
		"certificates": {
			"type": "array",
			"items": { "$ref": "#/$defs/X509Certtificate" }
		}
	},
	"required": ["service_provider_id", "certificates"]
}
`)

type SAMLSpSigningCertificate struct {
	ServiceProviderID string             `json:"service_provider_id,omitempty"`
	Certificates      []X509Certtificate `json:"certificates,omitempty"`
}
