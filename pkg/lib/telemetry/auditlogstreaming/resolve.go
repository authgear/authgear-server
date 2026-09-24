package auditlogstreaming

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// ResolvedStream is what the sender needs to deliver to one stream: parsed
// certificates and a resolved *x509.CertPool, rather than the raw config
// and secret documents.
type ResolvedStream struct {
	AppID string
	Name  string

	// Exactly one of Syslog and Datadog is non-nil: the encoding the
	// stream's type selects.
	Syslog  *ResolvedSyslog
	Datadog *ResolvedDatadog

	// Exactly one of TCP and HTTP is non-nil: the delivery its transport
	// selects.
	TCP  *ResolvedTCP
	HTTP *ResolvedHTTP
}

type ResolvedTCP struct {
	Address    string
	TLSEnabled bool
	// TLSClientCertificate is nil unless the stream declares client_certificate.
	TLSClientCertificate *tls.Certificate
	// TLSRootCAs is nil unless the stream declares certificate_authority,
	// in which case the system trust store is not consulted.
	TLSRootCAs *x509.CertPool
}

type ResolvedHTTP struct {
	// Endpoint is the absolute URL every chunk is POSTed to, already
	// derived from the encoding object when http.endpoint is unset.
	Endpoint string
	// RootCAs carries the tls secret's certificate_authority, for a
	// self-hosted endpoint with a private CA. nil means the system trust
	// store. client_certificate is not used by this transport.
	RootCAs *x509.CertPool
}

// resolveStream resolves a stream's config and secrets (if any) into a
// ResolvedStream. Part 01's config.SetFieldDefaults guarantees
// streamConfig.Syslog/TCP are non-nil for a syslog/tcp stream and
// streamConfig.Datadog/HTTP are non-nil for a datadog/http stream, so each
// branch below reads its own objects without nil guards, and must not
// read the other pair's, which is nil rather than an empty struct.
func resolveStream(
	appID string,
	streamConfig *config.TelemetryAuditLogStreamConfig,
	materials *config.TelemetryAuditLogStreamTLSMaterials,
	credentials *config.TelemetryAuditLogStreamDatadogCredentials,
) (*ResolvedStream, error) {
	resolved := &ResolvedStream{
		AppID: appID,
		Name:  streamConfig.Name,
	}

	// Looked up once, used by whichever transport branch runs. Resolve is
	// nil-safe.
	item, _ := materials.Resolve(streamConfig.Name)

	switch streamConfig.Type {
	case config.TelemetryAuditLogStreamTypeSyslog:
		resolved.Syslog = &ResolvedSyslog{
			Format:           streamConfig.Syslog.Format,
			Framing:          streamConfig.Syslog.Framing,
			Facility:         streamConfig.Syslog.Facility.Code(),
			AppName:          streamConfig.Syslog.AppName,
			StructuredDataID: streamConfig.Syslog.StructuredDataID,
		}
	case config.TelemetryAuditLogStreamTypeDatadog:
		datadog, err := resolveDatadog(streamConfig, credentials)
		if err != nil {
			return nil, err
		}
		resolved.Datadog = datadog
	default:
		return nil, fmt.Errorf("auditlogstreaming: unknown stream type %q for stream %q", streamConfig.Type, streamConfig.Name)
	}

	switch streamConfig.Transport {
	case config.TelemetryAuditLogStreamTransportTCP:
		tcp := &ResolvedTCP{
			Address:    streamConfig.TCP.Address,
			TLSEnabled: streamConfig.TCP.TLS.Enabled,
		}

		// An item pointing at a non-TLS stream is tolerated, not
		// rejected, by part 01's secret validation, so this branch is
		// skipped, not erroring, when TLS is disabled.
		if tcp.TLSEnabled && item != nil {
			if item.CertificateAuthority != nil {
				pool := x509.NewCertPool()
				if !pool.AppendCertsFromPEM([]byte(item.CertificateAuthority.Pem)) {
					return nil, fmt.Errorf("auditlogstreaming: invalid certificate_authority pem for stream %q", streamConfig.Name)
				}
				tcp.TLSRootCAs = pool
			}

			if item.ClientCertificate != nil {
				cert, err := parseClientCertificate(item.ClientCertificate)
				if err != nil {
					return nil, fmt.Errorf("auditlogstreaming: invalid client_certificate for stream %q: %w", streamConfig.Name, err)
				}
				tcp.TLSClientCertificate = cert
			}
		}

		resolved.TCP = tcp
	case config.TelemetryAuditLogStreamTransportHTTP:
		http, err := resolveHTTP(streamConfig, item)
		if err != nil {
			return nil, err
		}
		resolved.HTTP = http
	default:
		return nil, fmt.Errorf("auditlogstreaming: unknown stream transport %q for stream %q", streamConfig.Transport, streamConfig.Name)
	}

	return resolved, nil
}

// resolveDatadog resolves the datadog encoding object and its api key.
// The missing-key branch is defensive: part 01 §1.9 makes a keyless
// datadog stream fail to load, so a project reaching here has one. It
// returns an error rather than sending without the header, because an
// unauthenticated POST would be a 403 per chunk and a wasted round trip
// per tick.
func resolveDatadog(
	streamConfig *config.TelemetryAuditLogStreamConfig,
	credentials *config.TelemetryAuditLogStreamDatadogCredentials,
) (*ResolvedDatadog, error) {
	item, ok := credentials.Resolve(streamConfig.Name)
	if !ok || item.APIKey == "" {
		return nil, fmt.Errorf("auditlogstreaming: missing api key for datadog stream %q", streamConfig.Name)
	}

	// Sorted, because Go's map iteration order is randomised and an
	// unstable ddtags would make every golden test flaky and every log
	// line differ from the last for no reason. app_id and activity_type
	// come from the event, and the spec drops a configured pair that
	// would collide with them.
	keys := make([]string, 0, len(streamConfig.Datadog.Tags))
	for k := range streamConfig.Datadog.Tags {
		if k == "app_id" || k == "activity_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tags := make([]DatadogTag, 0, len(keys))
	for _, k := range keys {
		tags = append(tags, DatadogTag{Key: k, Value: streamConfig.Datadog.Tags[k]})
	}

	return &ResolvedDatadog{
		APIKey:  item.APIKey,
		Service: streamConfig.Datadog.Service,
		Source:  streamConfig.Datadog.Source,
		Tags:    tags,
	}, nil
}

// resolveHTTP resolves the http transport object: the endpoint, derived
// from the encoding object when http.endpoint is unset, and the tls
// secret's certificate_authority, when present.
func resolveHTTP(
	streamConfig *config.TelemetryAuditLogStreamConfig,
	item *config.TelemetryAuditLogStreamTLSMaterialsItem,
) (*ResolvedHTTP, error) {
	resolved := &ResolvedHTTP{}

	switch {
	case streamConfig.HTTP.Endpoint != "":
		resolved.Endpoint = streamConfig.HTTP.Endpoint
	case streamConfig.Datadog != nil:
		resolved.Endpoint = streamConfig.Datadog.Site.LogsIntakeURL()
	}

	// Unlike the tcp branch, there is no TLSEnabled gate: the endpoint's
	// scheme decides. Resolving a pool for a plaintext stream costs one
	// parse per tick and keeps this branch free of a scheme check that
	// would then have to agree with the one in net/http, which simply
	// never builds a TLS connection for an http:// URL.
	if item != nil && item.CertificateAuthority != nil {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(item.CertificateAuthority.Pem)) {
			return nil, fmt.Errorf("auditlogstreaming: invalid certificate_authority pem for stream %q", streamConfig.Name)
		}
		resolved.RootCAs = pool
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
