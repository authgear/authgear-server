package config

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

var _ = SecretConfigSchema.Add("X509Certificate", `
{
	"type": "object",
	"properties": {
		"pem": { "$ref": "#/$defs/X509CertificatePem" }
	},
	"required": ["pem"]
}
`)

type X509Certificate struct {
	Pem X509CertificatePem `json:"pem,omitempty"`
}

func (c *X509Certificate) Data() []byte {
	block, _ := pem.Decode([]byte(c.Pem))
	if block == nil {
		panic(fmt.Errorf("invalid pem"))
	}
	return block.Bytes
}

func (c *X509Certificate) Base64Data() string {
	return base64.StdEncoding.EncodeToString(c.Data())
}

func (c *X509Certificate) X509Certificate() *x509.Certificate {
	cert, err := x509.ParseCertificate(c.Data())
	if err != nil {
		panic(fmt.Errorf("failed to parse a stored X509Certificate: %w", err))
	}
	return cert
}

var _ = SecretConfigSchema.Add("X509CertificatePem", `
{
	"type": "string",
	"format": "x_x509_certificate_pem"
}
`)

type X509CertificatePem string
