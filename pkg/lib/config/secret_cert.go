package config

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

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

func (c *X509Cert) Base64Data() string {
	block, _ := pem.Decode([]byte(c.Pem))
	if block == nil {
		panic(fmt.Errorf("invalid pem"))
	}
	return base64.StdEncoding.EncodeToString(block.Bytes)
}

var _ = SecretConfigSchema.Add("X509CertPem", `
{
	"type": "string",
	"format": "x_x509_cert_pem"
}
`)

type X509CertPem string
