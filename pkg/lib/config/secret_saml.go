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
		"cert": { "$ref": "#/$defs/SAMLIdpSigningPemCert" },
		"key": { "$ref": "#/$defs/JWK" }
	},
	"required": ["cert", "key"]
}
`)

type SAMLIdpSigningCert struct {
	Cert *SAMLIdpSigningPemCert `json:"cert,omitempty"`
	Key  *JWK                   `json:"key,omitempty"`
}

var _ = SecretConfigSchema.Add("SAMLIdpSigningPemCert", `
{
	"type": "object",
	"properties": {
		"pem": { "$ref": "#/$defs/X509CertPem" }
	},
	"required": ["pem"]
}
`)

type SAMLIdpSigningPemCert struct {
	Pem X509CertPem `json:"pem,omitempty"`
}
