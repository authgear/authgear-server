package config_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const testCertificatePEM = "-----BEGIN CERTIFICATE-----\n" +
	"MIIDejCCAmKgAwIBAgIgLKKTB6GZMFHZVUiFIq8LcNIr0p8HFHwKM6r5/BQ/un4w\n" +
	"DQYJKoZIhvcNAQEFBQAwUDEJMAcGA1UEBhMAMQkwBwYDVQQKDAAxCTAHBgNVBAsM\n" +
	"ADENMAsGA1UEAwwEdGVzdDEPMA0GCSqGSIb3DQEJARYAMQ0wCwYDVQQDDAR0ZXN0\n" +
	"MB4XDTI0MDgwODA2NTY0OFoXDTM0MDgwOTA2NTY0OFowQTEJMAcGA1UEBhMAMQkw\n" +
	"BwYDVQQKDAAxCTAHBgNVBAsMADENMAsGA1UEAwwEdGVzdDEPMA0GCSqGSIb3DQEJ\n" +
	"ARYAMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5zRfTtkaa7cIsQS+\n" +
	"F1Dg25wPEvcjHsHcq598n+RzRJzfSLRtYwgEfs0VhyjHfo2O7KhNFh5cqdkEfzwA\n" +
	"bfxtgVLvy3yUjTMFO0FnJqrO3dkGiOAl654XUlXb4rF8DF1sPnUdd9QEZaZHGV/8\n" +
	"YuVOc3RV15jsr2jB9rra9//guAQ0CSP4XLJ5m9vf9nJILAHLryFIzDSgOVmhi4Ig\n" +
	"o59e9n3Hemavrta2C5Zj4cP6RNwuCV/i5lQOkzJIgksH9/EZCsR93DMEgkBS5oQQ\n" +
	"rt9Bzlr03TNGW4n/CYKNULK/osqJd5r5g3zUaQZY2KAan+oSsEXvBjzYtrehN1dm\n" +
	"dfbUEQIDAQABo08wTTAdBgNVHQ4EFgQUiXG6MG9PSB/clTIuzm8rW+8xLWkwHwYD\n" +
	"VR0jBBgwFoAUiXG6MG9PSB/clTIuzm8rW+8xLWkwCwYDVR0RBAQwAoIAMA0GCSqG\n" +
	"SIb3DQEBBQUAA4IBAQBTjdS9po3eEXukksMK6xBL3kQF1MEFUaWcgoN+h497lS9J\n" +
	"Xe1rmWpdZ1Aehp21GQmniRKU8uPLPRQKoX8Mhc/d3fHyv9u0YPns/2Wm8TBzxwHY\n" +
	"V2KdXZfpBdN+Z5bBRbgtKxx1z2GBfB39S2WCakS9xK8f7fuQPLIZz8eq7so5T8Hm\n" +
	"TU95acndEpnA0u6/MjbvXtZesTRZCewQw4CkcSLTCzB8dLG55UXHytnISWlCpuAx\n" +
	"8svq/ryZIi5vhBQFO/hG9s2Q32VvfKt2ZW8qA+gvOxEVDfAEFekKokP0Taiz77Q2\n" +
	"AVZxEXeABxJGtiMunQTr2q1tCrJQN0d08xlA5jXl\n" +
	"-----END CERTIFICATE-----\n"

const testJWKJSON = `{
	"kty": "RSA",
	"kid": "test",
	"n": "5zRfTtkaa7cIsQS-F1Dg25wPEvcjHsHcq598n-RzRJzfSLRtYwgEfs0VhyjHfo2O7KhNFh5cqdkEfzwAbfxtgVLvy3yUjTMFO0FnJqrO3dkGiOAl654XUlXb4rF8DF1sPnUdd9QEZaZHGV_8YuVOc3RV15jsr2jB9rra9__guAQ0CSP4XLJ5m9vf9nJILAHLryFIzDSgOVmhi4Igo59e9n3Hemavrta2C5Zj4cP6RNwuCV_i5lQOkzJIgksH9_EZCsR93DMEgkBS5oQQrt9Bzlr03TNGW4n_CYKNULK_osqJd5r5g3zUaQZY2KAan-oSsEXvBjzYtrehN1dmdfbUEQ",
	"e": "AQAB",
	"d": "AQAB"
}`

func TestTelemetryAuditLogStreamTLSMaterials(t *testing.T) {
	ctx := context.Background()

	parseSecrets := func(dataYAML string) (*config.TelemetryAuditLogStreamTLSMaterials, error) {
		yaml := "secrets:\n- key: telemetry.audit_logs.streams.tls\n  data:\n" + dataYAML
		secretConfig, err := config.ParseSecret(ctx, []byte(yaml))
		if err != nil {
			return nil, err
		}
		materials, _ := secretConfig.LookupData(config.TelemetryAuditLogStreamTLSMaterialsKey).(*config.TelemetryAuditLogStreamTLSMaterials)
		return materials, nil
	}

	Convey("TelemetryAuditLogStreamTLSMaterials", t, func() {
		Convey("client_certificate only parses", func() {
			_, err := parseSecrets(`
  - stream_name: collector
    client_certificate:
      certificate:
        pem: "` + escapeForYAML(testCertificatePEM) + `"
      key: ` + testJWKJSON + `
`)
			So(err, ShouldBeNil)
		})

		Convey("certificate_authority only parses", func() {
			_, err := parseSecrets(`
  - stream_name: collector
    certificate_authority:
      pem: "` + escapeForYAML(testCertificatePEM) + `"
`)
			So(err, ShouldBeNil)
		})

		Convey("both parse", func() {
			_, err := parseSecrets(`
  - stream_name: collector
    client_certificate:
      certificate:
        pem: "` + escapeForYAML(testCertificatePEM) + `"
      key: ` + testJWKJSON + `
    certificate_authority:
      pem: "` + escapeForYAML(testCertificatePEM) + `"
`)
			So(err, ShouldBeNil)
		})

		Convey("neither is rejected", func() {
			_, err := parseSecrets(`
  - stream_name: collector
`)
			So(err, ShouldBeError)
		})

		Convey("client_certificate with certificate but no key is rejected", func() {
			_, err := parseSecrets(`
  - stream_name: collector
    client_certificate:
      certificate:
        pem: "` + escapeForYAML(testCertificatePEM) + `"
`)
			So(err, ShouldBeError)
		})

		Convey("client_certificate with key but no certificate is rejected", func() {
			_, err := parseSecrets(`
  - stream_name: collector
    client_certificate:
      key: ` + testJWKJSON + `
`)
			So(err, ShouldBeError)
		})

		Convey("Resolve returns the matching item", func() {
			materials, err := parseSecrets(`
  - stream_name: collector
    certificate_authority:
      pem: "` + escapeForYAML(testCertificatePEM) + `"
`)
			So(err, ShouldBeNil)
			item, ok := materials.Resolve("collector")
			So(ok, ShouldBeTrue)
			So(item.StreamName, ShouldEqual, "collector")
		})

		Convey("Resolve returns false for an unknown name", func() {
			materials, err := parseSecrets(`
  - stream_name: collector
    certificate_authority:
      pem: "` + escapeForYAML(testCertificatePEM) + `"
`)
			So(err, ShouldBeNil)
			_, ok := materials.Resolve("unknown")
			So(ok, ShouldBeFalse)
		})

		Convey("Resolve on a nil receiver returns false", func() {
			var materials *config.TelemetryAuditLogStreamTLSMaterials
			_, ok := materials.Resolve("collector")
			So(ok, ShouldBeFalse)
		})
	})
}

func TestTelemetryAuditLogStreamDatadogCredentials(t *testing.T) {
	ctx := context.Background()

	parseSecrets := func(dataYAML string) (*config.TelemetryAuditLogStreamDatadogCredentials, error) {
		yaml := "secrets:\n- key: telemetry.audit_logs.streams.datadog\n  data:\n" + dataYAML
		secretConfig, err := config.ParseSecret(ctx, []byte(yaml))
		if err != nil {
			return nil, err
		}
		credentials, _ := secretConfig.LookupData(config.TelemetryAuditLogStreamDatadogCredentialsKey).(*config.TelemetryAuditLogStreamDatadogCredentials)
		return credentials, nil
	}

	Convey("TelemetryAuditLogStreamDatadogCredentials", t, func() {
		Convey("a valid item parses", func() {
			_, err := parseSecrets(`
  - stream_name: datadog
    api_key: "1234567890abcdef1234567890abcdef"
`)
			So(err, ShouldBeNil)
		})

		Convey("missing api_key is rejected", func() {
			_, err := parseSecrets(`
  - stream_name: datadog
`)
			So(err, ShouldBeError)
		})

		Convey("missing stream_name is rejected", func() {
			_, err := parseSecrets(`
  - api_key: "1234567890abcdef1234567890abcdef"
`)
			So(err, ShouldBeError)
		})

		Convey("empty api_key is rejected", func() {
			_, err := parseSecrets(`
  - stream_name: datadog
    api_key: ""
`)
			So(err, ShouldBeError)
		})

		Convey("unknown property is rejected", func() {
			_, err := parseSecrets(`
  - stream_name: datadog
    api_key: "1234567890abcdef1234567890abcdef"
    unknown: true
`)
			So(err, ShouldBeError)
		})

		Convey("Resolve returns the matching item", func() {
			credentials, err := parseSecrets(`
  - stream_name: datadog
    api_key: "1234567890abcdef1234567890abcdef"
`)
			So(err, ShouldBeNil)
			item, ok := credentials.Resolve("datadog")
			So(ok, ShouldBeTrue)
			So(item.APIKey, ShouldEqual, "1234567890abcdef1234567890abcdef")
		})

		Convey("Resolve returns false for an unknown name", func() {
			credentials, err := parseSecrets(`
  - stream_name: datadog
    api_key: "1234567890abcdef1234567890abcdef"
`)
			So(err, ShouldBeNil)
			_, ok := credentials.Resolve("unknown")
			So(ok, ShouldBeFalse)
		})

		Convey("Resolve on a nil receiver returns false", func() {
			var credentials *config.TelemetryAuditLogStreamDatadogCredentials
			_, ok := credentials.Resolve("datadog")
			So(ok, ShouldBeFalse)
		})

		Convey("SensitiveStrings returns every api_key", func() {
			credentials, err := parseSecrets(`
  - stream_name: datadog
    api_key: "key-one"
  - stream_name: datadog2
    api_key: "key-two"
`)
			So(err, ShouldBeNil)
			So(credentials.SensitiveStrings(), ShouldResemble, []string{"key-one", "key-two"})
		})

		Convey("SensitiveStrings on a nil receiver returns nil", func() {
			var credentials *config.TelemetryAuditLogStreamDatadogCredentials
			So(credentials.SensitiveStrings(), ShouldBeNil)
		})
	})
}

// escapeForYAML escapes a PEM block (which contains literal newlines) for
// embedding inside a double-quoted YAML scalar.
func escapeForYAML(s string) string {
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
