package config

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

var _ = SecretConfigSchema.Add("X509Certtificate", `
{
	"type": "object",
	"properties": {
		"pem": { "$ref": "#/$defs/X509CertificatePem" }
	},
	"required": ["pem"]
}
`)

type X509Certtificate struct {
	Pem X509CertificatePem `json:"pem,omitempty"`
}

func (c *X509Certtificate) Base64Data() string {
	block, _ := pem.Decode([]byte(c.Pem))
	if block == nil {
		panic(fmt.Errorf("invalid pem"))
	}
	return base64.StdEncoding.EncodeToString(block.Bytes)
}

var _ = SecretConfigSchema.Add("X509CertificatePem", `
{
	"type": "string",
	"format": "x_x509_certificate_pem"
}
`)

type X509CertificatePem string
