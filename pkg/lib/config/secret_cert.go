package config

import (
	// nolint:gosec
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
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

func (c *X509Certificate) Fingerprint() string {
	// nolint:gosec
	fingerprintBytes := sha1.Sum(c.X509Certificate().Raw)
	fingerprintHex := []string{}
	for _, b := range fingerprintBytes {
		fingerprintHex = append(fingerprintHex, fmt.Sprintf("%02X", b))
	}

	return strings.Join(fingerprintHex, ":")
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
