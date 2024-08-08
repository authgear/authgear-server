package config

var _ = SecretConfigSchema.Add("X509CertPem", `
{
	"type": "string",
	"format": "x_x509_cert_pem"
}
`)

type X509CertPem string
