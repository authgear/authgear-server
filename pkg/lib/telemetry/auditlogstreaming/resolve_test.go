package auditlogstreaming

import (
	"context"
	"crypto/x509"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// parseTLSMaterials parses materials through the real config.ParseSecret
// path, exactly as they arrive at runtime -- rather than a literal Go
// struct construction, which bypasses config.SetFieldDefaults entirely.
// That distinction matters here: SetFieldDefaults materializes
// ClientCertificate and CertificateAuthority into non-nil-but-empty
// structs whenever a document omits them (the same hazard part 01's
// cleanupOrphanedSecrets already had to work around), and a materials
// value built by a literal struct never exercises that path. A resolver
// bug that only manifests on defaulted-but-empty fields -- exactly the
// class of bug this package once had -- would pass every test built the
// other way.
func parseTLSMaterials(t *testing.T, dataYAML string) *config.TelemetryAuditLogStreamTLSMaterials {
	t.Helper()
	yaml := "secrets:\n- key: telemetry.audit_logs.streams.tls\n  data:\n" + dataYAML
	secretConfig, err := config.ParseSecret(context.Background(), []byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	materials, ok := secretConfig.LookupData(config.TelemetryAuditLogStreamTLSMaterialsKey).(*config.TelemetryAuditLogStreamTLSMaterials)
	if !ok {
		t.Fatal("expected *config.TelemetryAuditLogStreamTLSMaterials")
	}
	return materials
}

func newTestStreamConfig(name string, tlsEnabled bool) *config.TelemetryAuditLogStreamConfig {
	return &config.TelemetryAuditLogStreamConfig{
		Name:      name,
		Type:      config.TelemetryAuditLogStreamTypeSyslog,
		Transport: config.TelemetryAuditLogStreamTransportTCP,
		TCP: &config.TelemetryAuditLogStreamTCPConfig{
			Address: "collector.internal:5140",
			TLS:     &config.TelemetryAuditLogStreamTCPTLSConfig{Enabled: tlsEnabled},
		},
		Syslog: &config.TelemetryAuditLogStreamSyslogConfig{
			Format:           config.SyslogFormatRFC5424,
			Framing:          config.SyslogFramingNewline,
			Facility:         config.SyslogFacilityLocal0,
			AppName:          "authgear",
			StructuredDataID: "authgear",
		},
	}
}

func TestResolveStream(t *testing.T) {
	Convey("resolveStream", t, func() {
		ca := newTestCA(t, "test-ca")

		Convey("TLS disabled: no TLS material even when a matching secret item exists", func() {
			streamConfig := newTestStreamConfig("collector", false)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: ca.pem()}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeNil)
			So(resolved.TLSEnabled, ShouldBeFalse)
			So(resolved.TLSRootCAs, ShouldBeNil)
			So(resolved.TLSClientCertificate, ShouldBeNil)
		})

		Convey("TLS enabled, no secret item: both nil", func() {
			streamConfig := newTestStreamConfig("collector", true)
			var materials config.TelemetryAuditLogStreamTLSMaterials
			resolved, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeNil)
			So(resolved.TLSRootCAs, ShouldBeNil)
			So(resolved.TLSClientCertificate, ShouldBeNil)
		})

		Convey("certificate_authority only: TLSRootCAs non-nil, TLSClientCertificate nil", func() {
			streamConfig := newTestStreamConfig("collector", true)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: ca.pem()}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeNil)
			So(resolved.TLSRootCAs, ShouldNotBeNil)
			So(resolved.TLSClientCertificate, ShouldBeNil)
		})

		Convey("client_certificate only: the reverse", func() {
			streamConfig := newTestStreamConfig("collector", true)
			l := ca.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
					Certificate: &config.X509Certificate{Pem: l.certPEM},
					Key:         l.jwk,
				}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeNil)
			So(resolved.TLSRootCAs, ShouldBeNil)
			So(resolved.TLSClientCertificate, ShouldNotBeNil)
		})

		Convey("both: both", func() {
			streamConfig := newTestStreamConfig("collector", true)
			l := ca.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{
					StreamName:           "collector",
					CertificateAuthority: &config.X509Certificate{Pem: ca.pem()},
					ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
						Certificate: &config.X509Certificate{Pem: l.certPEM},
						Key:         l.jwk,
					},
				},
			}
			resolved, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeNil)
			So(resolved.TLSRootCAs, ShouldNotBeNil)
			So(resolved.TLSClientCertificate, ShouldNotBeNil)
		})

		Convey("a certificate chain with a leaf plus one intermediate produces two chain entries", func() {
			streamConfig := newTestStreamConfig("collector", true)
			l := ca.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)
			chainPEM := string(l.certPEM) + string(ca.pem())
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
					Certificate: &config.X509Certificate{Pem: config.X509CertificatePem(chainPEM)},
					Key:         l.jwk,
				}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeNil)
			So(resolved.TLSClientCertificate.Certificate, ShouldHaveLength, 2)
		})

		Convey("malformed CA PEM returns an error rather than panicking", func() {
			streamConfig := newTestStreamConfig("collector", true)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: "not a pem"}},
			}
			_, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeError)
		})

		Convey("regression: certificate_authority only, parsed through the real config.ParseSecret path", func() {
			streamConfig := newTestStreamConfig("collector", true)
			materials := parseTLSMaterials(t, `
  - stream_name: collector
    certificate_authority:
      pem: "`+escapeYAMLString(string(ca.pem()))+`"
`)
			resolved, err := resolveStream("app", streamConfig, materials)
			So(err, ShouldBeNil)
			So(resolved.TLSRootCAs, ShouldNotBeNil)
			So(resolved.TLSClientCertificate, ShouldBeNil)
		})

		Convey("regression: client_certificate only, parsed through the real config.ParseSecret path", func() {
			streamConfig := newTestStreamConfig("collector", true)
			l := ca.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)
			materials := parseTLSMaterials(t, `
  - stream_name: collector
    client_certificate:
      certificate:
        pem: "`+escapeYAMLString(string(l.certPEM))+`"
      key: `+l.jwkJSON+`
`)
			resolved, err := resolveStream("app", streamConfig, materials)
			So(err, ShouldBeNil)
			So(resolved.TLSRootCAs, ShouldBeNil)
			So(resolved.TLSClientCertificate, ShouldNotBeNil)
		})

		Convey("a JWK that does not match the certificate returns an error rather than panicking", func() {
			streamConfig := newTestStreamConfig("collector", true)
			l := ca.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)
			other := newTestCA(t, "other")
			otherLeaf := other.issueLeaf(t, "other-client", x509.ExtKeyUsageClientAuth, nil, nil)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
					Certificate: &config.X509Certificate{Pem: l.certPEM},
					// A syntactically valid JWK, but not the key for l's certificate.
					Key: otherLeaf.jwk,
				}},
			}
			_, err := resolveStream("app", streamConfig, &materials)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "does not match")
		})
	})
}
