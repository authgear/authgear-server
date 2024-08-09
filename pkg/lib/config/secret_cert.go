package config

var _ = SecretConfigSchema.Add("X509Cert", `
{
	"type": "object",
	"properties": {
		"pem": { "$ref": "#/$defs/X509CertPem" }
	},
	"required": ["pem"]
}
`)

type X509Cert struct {
	Pem X509CertPem `json:"pem,omitempty"`
}

var _ = SecretConfigSchema.Add("X509CertPem", `
{
	"type": "string",
	"format": "x_x509_cert_pem"
}
`)

type X509CertPem string
