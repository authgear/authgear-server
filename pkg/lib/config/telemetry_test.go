package config_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const telemetryBaseYAML = `
id: test
http:
  public_origin: http://test
`

func parseTelemetryStreams(streamsYAML string) (*config.AppConfig, error) {
	ctx := context.Background()
	yaml := telemetryBaseYAML + "telemetry:\n  audit_logs:\n    streams:\n" + streamsYAML
	return config.Parse(ctx, []byte(yaml))
}

func TestTelemetryAuditLogStreamDefaults(t *testing.T) {
	Convey("TelemetryAuditLogStreamConfig defaults", t, func() {
		Convey("minimal valid stream defaults facility, app_name, structured_data_id and tls.enabled", func() {
			cfg, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
`)
			So(err, ShouldBeNil)
			So(cfg.Telemetry.AuditLogs.Streams, ShouldHaveLength, 1)
			stream := cfg.Telemetry.AuditLogs.Streams[0]
			So(stream.Syslog.Facility, ShouldEqual, config.SyslogFacilityLocal0)
			So(stream.Syslog.AppName, ShouldEqual, "authgear")
			So(stream.Syslog.StructuredDataID, ShouldEqual, "authgear")
			So(stream.TCP.TLS.Enabled, ShouldBeFalse)
		})

		Convey("explicit facility, app_name, structured_data_id survive defaulting", func() {
			cfg, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
        facility: local3
        app_name: myapp
        structured_data_id: myapp@12345
`)
			So(err, ShouldBeNil)
			stream := cfg.Telemetry.AuditLogs.Streams[0]
			So(stream.Syslog.Facility, ShouldEqual, config.SyslogFacilityLocal3)
			So(stream.Syslog.AppName, ShouldEqual, "myapp")
			So(stream.Syslog.StructuredDataID, ShouldEqual, "myapp@12345")
		})
	})
}

func TestTelemetryAuditLogStreamDatadogDefaults(t *testing.T) {
	Convey("TelemetryAuditLogStreamDatadogConfig defaults", t, func() {
		Convey("minimal datadog stream defaults site, service and source, and leaves tags nil", func() {
			cfg, err := parseTelemetryStreams(`
    - name: datadog
      type: datadog
      transport: http
`)
			So(err, ShouldBeNil)
			So(cfg.Telemetry.AuditLogs.Streams, ShouldHaveLength, 1)
			stream := cfg.Telemetry.AuditLogs.Streams[0]
			So(stream.Datadog.Site, ShouldEqual, config.DatadogSiteUS1)
			So(stream.Datadog.Service, ShouldEqual, "authgear")
			So(stream.Datadog.Source, ShouldEqual, "authgear")
			So(stream.Datadog.Tags, ShouldBeNil)
		})

		Convey("explicit site, service, source and tags survive defaulting", func() {
			cfg, err := parseTelemetryStreams(`
    - name: datadog
      type: datadog
      transport: http
      datadog:
        site: datadoghq.eu
        service: myservice
        source: myservice
        tags:
          env: production
`)
			So(err, ShouldBeNil)
			stream := cfg.Telemetry.AuditLogs.Streams[0]
			So(stream.Datadog.Site, ShouldEqual, config.DatadogSite("datadoghq.eu"))
			So(stream.Datadog.Service, ShouldEqual, "myservice")
			So(stream.Datadog.Source, ShouldEqual, "myservice")
			So(stream.Datadog.Tags, ShouldResemble, map[string]string{"env": "production"})
		})

		Convey("a datadog stream has Syslog and TCP nil after parsing", func() {
			cfg, err := parseTelemetryStreams(`
    - name: datadog
      type: datadog
      transport: http
`)
			So(err, ShouldBeNil)
			stream := cfg.Telemetry.AuditLogs.Streams[0]
			So(stream.Syslog, ShouldBeNil)
			So(stream.TCP, ShouldBeNil)
			So(stream.HTTP, ShouldNotBeNil)
			So(stream.Datadog, ShouldNotBeNil)
		})

		Convey("a syslog stream has Datadog and HTTP nil after parsing", func() {
			cfg, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
`)
			So(err, ShouldBeNil)
			stream := cfg.Telemetry.AuditLogs.Streams[0]
			So(stream.Datadog, ShouldBeNil)
			So(stream.HTTP, ShouldBeNil)
			So(stream.Syslog, ShouldNotBeNil)
			So(stream.TCP, ShouldNotBeNil)
		})
	})
}

func TestDatadogSiteLogsIntakeURL(t *testing.T) {
	Convey("DatadogSite.LogsIntakeURL", t, func() {
		So(config.DatadogSite("datadoghq.eu").LogsIntakeURL(), ShouldEqual, "https://http-intake.logs.datadoghq.eu/api/v2/logs")
	})
}

func TestTelemetryAuditLogStreamRequiredBlocks(t *testing.T) {
	Convey("TelemetryAuditLogStreamConfig required blocks", t, func() {
		Convey("type: syslog without a syslog block is rejected", func() {
			_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
`)
			So(err, ShouldBeError)
		})

		Convey("transport: tcp without a tcp block is rejected", func() {
			_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      syslog:
        format: rfc5424
        framing: newline
`)
			So(err, ShouldBeError)
		})
	})
}

func TestTelemetryAuditLogStreamName(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty rejected", "", true},
		{"64 chars rejected", strings.Repeat("a", 64), true},
		{"dot rejected", "a.b", true},
		{"63 chars accepted", strings.Repeat("a", 63), false},
		{"hyphen and underscore accepted", "a-b_c", false},
	}

	Convey("TelemetryAuditLogStreamConfig.name", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				_, err := parseTelemetryStreams(`
    - name: "` + tc.value + `"
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
`)
				assertWantErr(tc.wantErr, err)
			})
		}

		Convey("duplicated stream names are rejected", func() {
			_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: other.internal:5140
      syslog:
        format: rfc5424
        framing: newline
`)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "duplicated audit log stream name")
		})

		Convey("distinct stream names are accepted", func() {
			_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
    - name: staging
      type: syslog
      transport: tcp
      tcp:
        address: other.internal:5140
      syslog:
        format: rfc5424
        framing: newline
`)
			So(err, ShouldBeNil)
		})
	})
}

func TestTelemetryAuditLogStreamSyslogFacility(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"unknown value rejected", "local8", true},
		{"the old numeric form is rejected", "16", true},
		{"user accepted", "user", false},
		{"local0 accepted", "local0", false},
		{"local1 accepted", "local1", false},
		{"local2 accepted", "local2", false},
		{"local3 accepted", "local3", false},
		{"local4 accepted", "local4", false},
		{"local5 accepted", "local5", false},
		{"local6 accepted", "local6", false},
		{"local7 accepted", "local7", false},
	}

	Convey("TelemetryAuditLogStreamSyslogConfig.facility", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
        facility: ` + tc.value + `
`)
				assertWantErr(tc.wantErr, err)
			})
		}
	})
}

func TestTelemetryAuditLogStreamSyslogFraming(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"unknown value rejected", "invalid", true},
		{"octet_counting accepted", "octet_counting", false},
		{"newline accepted", "newline", false},
	}

	Convey("TelemetryAuditLogStreamSyslogConfig.framing", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: ` + tc.value + `
`)
				assertWantErr(tc.wantErr, err)
			})
		}
	})
}

func TestTelemetryAuditLogStreamSyslogAppName(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"49 chars rejected", strings.Repeat("a", 49), true},
		{"space rejected", "my app", true},
		{"48 chars accepted", strings.Repeat("a", 48), false},
	}

	Convey("TelemetryAuditLogStreamSyslogConfig.app_name", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
        app_name: "` + tc.value + `"
`)
				assertWantErr(tc.wantErr, err)
			})
		}
	})
}

func TestTelemetryAuditLogStreamSyslogStructuredDataID(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"equals sign rejected", "a=b", true},
		{"closing bracket rejected", "a]b", true},
		{"quote rejected", `a"b`, true},
		{"space rejected", "a b", true},
		// This is the case that fails if the backslash in the pattern
		// regex is under-escaped: a single "\\" would make "^" mean an
		// escaped caret instead of the start of the "^-~" range, and
		// wrongly reject this otherwise-compliant value.
		{"caret accepted", "a^b", false},
	}

	Convey("TelemetryAuditLogStreamSyslogConfig.structured_data_id", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
        structured_data_id: "` + tc.value + `"
`)
				assertWantErr(tc.wantErr, err)
			})
		}
	})
}

func TestTelemetryAuditLogStreamTCPAddress(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"no port rejected", "collector.internal", true},
		{"no host rejected", ":5140", true},
		{"port zero rejected", "collector.internal:0", true},
		{"port too large rejected", "collector.internal:70000", true},
		{"hostname accepted", "collector.internal:5140", false},
		{"ipv4 accepted", "127.0.0.1:5140", false},
		{"ipv6 accepted", "[::1]:5140", false},
	}

	Convey("TelemetryAuditLogStreamTCPConfig.address", t, func() {
		for _, tc := range cases {
			Convey(tc.name, func() {
				_, err := parseTelemetryStreams(`
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: "` + tc.value + `"
      syslog:
        format: rfc5424
        framing: newline
`)
				assertWantErr(tc.wantErr, err)
			})
		}
	})
}

func assertWantErr(wantErr bool, err error) {
	if wantErr {
		So(err, ShouldBeError)
	} else {
		So(err, ShouldBeNil)
	}
}
