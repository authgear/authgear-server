package config

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

func GenerateSAMLIdpSigningCertificate(commonName string) (*SAMLIdpSigningCertificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		return nil, err
	}
	thumbprint, err := jwkKey.Thumbprint(crypto.SHA256)
	if err != nil {
		return nil, err
	}

	_ = jwkKey.Set("kid", base64.RawURLEncoding.EncodeToString(thumbprint))

	now := time.Now().UTC()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	tpl := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    now,
		NotAfter:     now.Add(50 * 365 * 24 * time.Hour), // 50 years
		KeyUsage:     x509.KeyUsageDigitalSignature,
		Subject: pkix.Name{
			CommonName: commonName,
		},
	}

	pubKey := &key.PublicKey
	certBytes, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pubKey, key)
	if err != nil {
		return nil, err
	}
	pemBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}

	var pemBuffer bytes.Buffer
	err = pem.Encode(&pemBuffer, &pemBlock)
	if err != nil {
		return nil, err
	}

	signingSecret := &SAMLIdpSigningCertificate{
		Certificate: &X509Certificate{
			Pem: X509CertificatePem(pemBuffer.String()),
		},
		Key: &JWK{
			Key: jwkKey,
		},
	}

	return signingSecret, nil
}
