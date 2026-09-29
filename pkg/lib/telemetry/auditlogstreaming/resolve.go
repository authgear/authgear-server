package auditlogstreaming

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// ResolvedStream is what the sender needs to deliver to one stream: parsed
// certificates and a resolved *x509.CertPool, rather than the raw config
// and secret documents.
type ResolvedStream struct {
	AppID string
	Name  string

	Address    string
	TLSEnabled bool
	// TLSClientCertificate is nil unless the stream declares client_certificate.
	TLSClientCertificate *tls.Certificate
	// TLSRootCAs is nil unless the stream declares certificate_authority,
	// in which case the system trust store is not consulted.
	TLSRootCAs *x509.CertPool

	Syslog ResolvedSyslog
}

// resolveStream resolves a stream's config and TLS secret (if any) into a
// ResolvedStream. Both streamConfig.TCP and streamConfig.TCP.TLS are
// non-nil after config.SetFieldDefaults (part 01), so they are read here
// without nil guards.
func resolveStream(
	appID string,
	streamConfig *config.TelemetryAuditLogStreamConfig,
	materials *config.TelemetryAuditLogStreamTLSMaterials,
) (*ResolvedStream, error) {
	resolved := &ResolvedStream{
		AppID:      appID,
		Name:       streamConfig.Name,
		Address:    streamConfig.TCP.Address,
		TLSEnabled: streamConfig.TCP.TLS.Enabled,
		Syslog: ResolvedSyslog{
			Format:           streamConfig.Syslog.Format,
			Framing:          streamConfig.Syslog.Framing,
			Facility:         streamConfig.Syslog.Facility.Code(),
			AppName:          streamConfig.Syslog.AppName,
			StructuredDataID: streamConfig.Syslog.StructuredDataID,
		},
	}

	// An item pointing at a non-TLS stream is tolerated, not rejected, by
	// part 01's secret validation, so this branch is reachable in a
	// healthy project.
	if !resolved.TLSEnabled {
		return resolved, nil
	}

	item, ok := materials.Resolve(streamConfig.Name)
	if !ok {
		// No secret item: verify against the system trust store, present
		// no client certificate.
		return resolved, nil
	}

	if item.CertificateAuthority != nil {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(item.CertificateAuthority.Pem)) {
			return nil, fmt.Errorf("auditlogstreaming: invalid certificate_authority pem for stream %q", streamConfig.Name)
		}
		resolved.TLSRootCAs = pool
	}

	if item.ClientCertificate != nil {
		cert, err := parseClientCertificate(item.ClientCertificate)
		if err != nil {
			return nil, fmt.Errorf("auditlogstreaming: invalid client_certificate for stream %q: %w", streamConfig.Name, err)
		}
		resolved.TLSClientCertificate = cert
	}

	return resolved, nil
}

// parseClientCertificate builds a tls.Certificate from a
// TelemetryAuditLogStreamClientCertificate: the chain is every PEM block
// in certificate.pem (leaf first, then intermediates), and the private key
// comes from the JWK. Errors are returned, not panicked -- unlike
// X509Certificate.Data(), which panics -- because this input comes from a
// secret file and a bad value must degrade the stream, not the worker.
func parseClientCertificate(cc *config.TelemetryAuditLogStreamClientCertificate) (*tls.Certificate, error) {
	if cc.Certificate == nil {
		return nil, fmt.Errorf("missing certificate")
	}
	if cc.Key == nil {
		return nil, fmt.Errorf("missing key")
	}

	var cert tls.Certificate
	rest := []byte(cc.Certificate.Pem)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		cert.Certificate = append(cert.Certificate, block.Bytes)
	}
	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("no certificate found in pem")
	}

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse leaf certificate: %w", err)
	}
	cert.Leaf = leaf

	var privateKey any
	if err := cc.Key.Key.Raw(&privateKey); err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	cert.PrivateKey = privateKey

	// Cross-check the private key against the leaf's public key, the same
	// way tls.X509KeyPair does for its PEM inputs. Extracting a raw JWK
	// never fails on its own even when the key belongs to a different
	// certificate entirely, so without this check a mismatched pair would
	// only surface much later, as an opaque handshake failure.
	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not support signing")
	}
	pub, ok := leaf.PublicKey.(interface{ Equal(x crypto.PublicKey) bool })
	if !ok {
		return nil, fmt.Errorf("unsupported certificate public key type %T", leaf.PublicKey)
	}
	if !pub.Equal(signer.Public()) {
		return nil, fmt.Errorf("private key does not match certificate public key")
	}

	return &cert, nil
}
