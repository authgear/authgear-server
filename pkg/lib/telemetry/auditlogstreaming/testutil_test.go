package auditlogstreaming

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// testCA is a self-signed certificate authority generated fresh for one
// test, used to sign leaf certificates.
type testCA struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
	der  []byte
}

func newTestCA(t *testing.T, commonName string) *testCA {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &testCA{cert: cert, key: key, der: der}
}

func (ca *testCA) pem() config.X509CertificatePem {
	return config.X509CertificatePem(pemEncode("CERTIFICATE", ca.der))
}

// pool returns an *x509.CertPool trusting only this CA, for a test
// listener's ClientCAs.
func (ca *testCA) pool() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)
	return pool
}

// leaf describes a certificate signed by the CA, plus its private key both
// as a config.JWK (for a TelemetryAuditLogStreamClientCertificate) and as
// a raw *rsa.PrivateKey (for building a plain tls.Certificate to serve
// from a test listener).
type leaf struct {
	certPEM config.X509CertificatePem
	der     []byte
	key     *rsa.PrivateKey
	jwk     *config.JWK
	jwkJSON string
}

// tlsCertificate builds a tls.Certificate directly from the leaf, for use
// as a test server's own certificate.
func (l *leaf) tlsCertificate() tls.Certificate {
	return tls.Certificate{
		Certificate: [][]byte{l.der},
		PrivateKey:  l.key,
	}
}

func (ca *testCA) issueLeaf(t *testing.T, commonName string, extKeyUsage x509.ExtKeyUsage, ipAddresses []net.IP, dnsNames []string) *leaf {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{extKeyUsage},
		IPAddresses:  ipAddresses,
		DNSNames:     dnsNames,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatal(err)
	}

	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, "test"); err != nil {
		t.Fatal(err)
	}
	jwkJSON, err := json.Marshal(jwkKey)
	if err != nil {
		t.Fatal(err)
	}
	var cfgJWK config.JWK
	if err := cfgJWK.UnmarshalJSON(jwkJSON); err != nil {
		t.Fatal(err)
	}

	return &leaf{
		certPEM: config.X509CertificatePem(pemEncode("CERTIFICATE", der)),
		der:     der,
		key:     key,
		jwk:     &cfgJWK,
		jwkJSON: string(jwkJSON),
	}
}

func pemEncode(blockType string, der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}))
}

// escapeYAMLString escapes a PEM block (which contains literal newlines)
// for embedding inside a double-quoted YAML scalar.
func escapeYAMLString(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			out.WriteString(`\n`)
		case '"':
			out.WriteString(`\"`)
		default:
			out.WriteString(string(r))
		}
	}
	return out.String()
}
