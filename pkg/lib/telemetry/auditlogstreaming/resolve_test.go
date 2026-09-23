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

func newTestDatadogStreamConfig(name string, datadogCfg *config.TelemetryAuditLogStreamDatadogConfig, httpEndpoint string) *config.TelemetryAuditLogStreamConfig {
	if datadogCfg == nil {
		datadogCfg = &config.TelemetryAuditLogStreamDatadogConfig{Site: config.DatadogSiteUS1}
	}
	return &config.TelemetryAuditLogStreamConfig{
		Name:      name,
		Type:      config.TelemetryAuditLogStreamTypeDatadog,
		Transport: config.TelemetryAuditLogStreamTransportHTTP,
		Datadog:   datadogCfg,
		HTTP:      &config.TelemetryAuditLogStreamHTTPConfig{Endpoint: httpEndpoint},
	}
}

func parseDatadogCredentials(t *testing.T, dataYAML string) *config.TelemetryAuditLogStreamDatadogCredentials {
	t.Helper()
	yaml := "secrets:\n- key: telemetry.audit_logs.streams.datadog\n  data:\n" + dataYAML
	secretConfig, err := config.ParseSecret(context.Background(), []byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	credentials, ok := secretConfig.LookupData(config.TelemetryAuditLogStreamDatadogCredentialsKey).(*config.TelemetryAuditLogStreamDatadogCredentials)
	if !ok {
		t.Fatal("expected *config.TelemetryAuditLogStreamDatadogCredentials")
	}
	return credentials
}

func TestResolveStreamDatadog(t *testing.T) {
	Convey("resolveStream: datadog/http", t, func() {
		credentials := &config.TelemetryAuditLogStreamDatadogCredentials{
			{StreamName: "datadog", APIKey: "e2e-datadog-api-key"},
		}

		Convey("a datadog stream resolves Datadog non-nil and Syslog nil, HTTP non-nil and TCP nil", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			resolved, err := resolveStream("app", streamConfig, nil, credentials)
			So(err, ShouldBeNil)
			So(resolved.Datadog, ShouldNotBeNil)
			So(resolved.Syslog, ShouldBeNil)
			So(resolved.HTTP, ShouldNotBeNil)
			So(resolved.TCP, ShouldBeNil)
		})

		Convey("a syslog stream resolves the reverse", func() {
			streamConfig := newTestStreamConfig("collector", false)
			resolved, err := resolveStream("app", streamConfig, nil, nil)
			So(err, ShouldBeNil)
			So(resolved.Syslog, ShouldNotBeNil)
			So(resolved.Datadog, ShouldBeNil)
			So(resolved.TCP, ShouldNotBeNil)
			So(resolved.HTTP, ShouldBeNil)
		})

		Convey("no matching credentials item returns an error", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			_, err := resolveStream("app", streamConfig, nil, &config.TelemetryAuditLogStreamDatadogCredentials{})
			So(err, ShouldBeError)
		})

		Convey("an item with an empty api_key returns an error", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			emptyKeyCredentials := &config.TelemetryAuditLogStreamDatadogCredentials{
				{StreamName: "datadog", APIKey: ""},
			}
			_, err := resolveStream("app", streamConfig, nil, emptyKeyCredentials)
			So(err, ShouldBeError)
		})

		Convey("resolveHTTP derives the intake URL from site, for the default site and an arbitrary one", func() {
			// site is not a closed enum (part 01 §1.2), so this proves the
			// URL interpolation, not a fixed list -- a site this package
			// has never heard of resolves the same way the default does.
			sites := []config.DatadogSite{
				config.DatadogSiteUS1,
				config.DatadogSite("eu1.datadoghq.com"),
			}
			for _, site := range sites {
				streamConfig := newTestDatadogStreamConfig("datadog", &config.TelemetryAuditLogStreamDatadogConfig{Site: site}, "")
				resolved, err := resolveStream("app", streamConfig, nil, credentials)
				So(err, ShouldBeNil)
				So(resolved.HTTP.Endpoint, ShouldEqual, site.LogsIntakeURL())
			}
		})

		Convey("resolveHTTP returns http.endpoint verbatim when it is set", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			resolved, err := resolveStream("app", streamConfig, nil, credentials)
			So(err, ShouldBeNil)
			So(resolved.HTTP.Endpoint, ShouldEqual, "https://opw.internal:8282/api/v2/logs")
		})

		Convey("certificate_authority in the tls secret becomes ResolvedHTTP.RootCAs", func() {
			ca := newTestCA(t, "test-ca")
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "datadog", CertificateAuthority: &config.X509Certificate{Pem: ca.pem()}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials, credentials)
			So(err, ShouldBeNil)
			So(resolved.HTTP.RootCAs, ShouldNotBeNil)
		})

		Convey("a malformed PEM in certificate_authority is an error", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "datadog", CertificateAuthority: &config.X509Certificate{Pem: "not a pem"}},
			}
			_, err := resolveStream("app", streamConfig, &materials, credentials)
			So(err, ShouldBeError)
		})

		Convey("client_certificate alone leaves RootCAs nil", func() {
			ca := newTestCA(t, "test-ca")
			l := ca.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "datadog", ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
					Certificate: &config.X509Certificate{Pem: l.certPEM},
					Key:         l.jwk,
				}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials, credentials)
			So(err, ShouldBeNil)
			So(resolved.HTTP.RootCAs, ShouldBeNil)
		})

		Convey("configured tags are sorted by key, and a configured app_id or activity_type pair is dropped", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", &config.TelemetryAuditLogStreamDatadogConfig{
				Site: config.DatadogSiteUS1,
				Tags: map[string]string{
					"team":          "security",
					"env":           "production",
					"app_id":        "should-be-dropped",
					"activity_type": "should-be-dropped",
				},
			}, "")
			resolved, err := resolveStream("app", streamConfig, nil, credentials)
			So(err, ShouldBeNil)
			So(resolved.Datadog.Tags, ShouldResemble, []DatadogTag{
				{Key: "env", Value: "production"},
				{Key: "team", Value: "security"},
			})
		})

		Convey("regression: credentials parsed through the real config.ParseSecret path", func() {
			streamConfig := newTestDatadogStreamConfig("datadog", nil, "https://opw.internal:8282/api/v2/logs")
			parsedCredentials := parseDatadogCredentials(t, `
  - stream_name: datadog
    api_key: "e2e-datadog-api-key"
`)
			resolved, err := resolveStream("app", streamConfig, nil, parsedCredentials)
			So(err, ShouldBeNil)
			So(resolved.Datadog.APIKey, ShouldEqual, "e2e-datadog-api-key")
		})
	})
}

func TestResolveStream(t *testing.T) {
	Convey("resolveStream", t, func() {
		ca := newTestCA(t, "test-ca")

		Convey("TLS disabled: no TLS material even when a matching secret item exists", func() {
			streamConfig := newTestStreamConfig("collector", false)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: ca.pem()}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSEnabled, ShouldBeFalse)
			So(resolved.TCP.TLSRootCAs, ShouldBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldBeNil)
		})

		Convey("TLS enabled, no secret item: both nil", func() {
			streamConfig := newTestStreamConfig("collector", true)
			var materials config.TelemetryAuditLogStreamTLSMaterials
			resolved, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSRootCAs, ShouldBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldBeNil)
		})

		Convey("certificate_authority only: TLSRootCAs non-nil, TLSClientCertificate nil", func() {
			streamConfig := newTestStreamConfig("collector", true)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: ca.pem()}},
			}
			resolved, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSRootCAs, ShouldNotBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldBeNil)
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
			resolved, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSRootCAs, ShouldBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldNotBeNil)
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
			resolved, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSRootCAs, ShouldNotBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldNotBeNil)
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
			resolved, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSClientCertificate.Certificate, ShouldHaveLength, 2)
		})

		Convey("malformed CA PEM returns an error rather than panicking", func() {
			streamConfig := newTestStreamConfig("collector", true)
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: "not a pem"}},
			}
			_, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeError)
		})

		Convey("regression: certificate_authority only, parsed through the real config.ParseSecret path", func() {
			streamConfig := newTestStreamConfig("collector", true)
			materials := parseTLSMaterials(t, `
  - stream_name: collector
    certificate_authority:
      pem: "`+escapeYAMLString(string(ca.pem()))+`"
`)
			resolved, err := resolveStream("app", streamConfig, materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSRootCAs, ShouldNotBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldBeNil)
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
			resolved, err := resolveStream("app", streamConfig, materials, nil)
			So(err, ShouldBeNil)
			So(resolved.TCP.TLSRootCAs, ShouldBeNil)
			So(resolved.TCP.TLSClientCertificate, ShouldNotBeNil)
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
			_, err := resolveStream("app", streamConfig, &materials, nil)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "does not match")
		})
	})
}
