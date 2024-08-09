package cmdinternal

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var cmdInternalSaml = &cobra.Command{
	Use:   "saml",
	Short: "SAML commands",
}

var cmdInternalSamlGenerateSigningKey = &cobra.Command{
	Use:   "generate-signing-key { common-name }",
	Short: "Generate a signing key with a x.509 cert",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("common-name is required")
		}
		commonName := args[0]

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		jwkKey, err := jwk.FromRaw(key)
		if err != nil {
			return err
		}
		thumbprint, err := jwkKey.Thumbprint(crypto.SHA256)
		if err != nil {
			return err
		}

		_ = jwkKey.Set("kid", base64.RawURLEncoding.EncodeToString(thumbprint))

		now := time.Now().UTC()

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			return err
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
			return err
		}
		pemBlock := pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		}

		var pemBuffer bytes.Buffer
		err = pem.Encode(&pemBuffer, &pemBlock)
		if err != nil {
			return err
		}

		signingSecret := config.SAMLIdpSigningCert{
			Cert: &config.X509Cert{
				Pem: config.X509CertPem(pemBuffer.String()),
			},
			Key: &config.JWK{
				Key: jwkKey,
			},
		}

		jsonBytes, err := json.Marshal(signingSecret)
		if err != nil {
			return err
		}

		yamlBytes, err := yaml.JSONToYAML(jsonBytes)
		if err != nil {
			return err
		}

		fmt.Println(string(yamlBytes))
		return nil
	},
}
