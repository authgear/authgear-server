package model

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestNewSecretConfigTelemetry(t *testing.T) {
	Convey("NewSecretConfig telemetry audit log stream secrets", t, func() {
		ctx := context.Background()
		cfg, err := config.ParsePartialSecret(ctx, []byte(`
secrets:
- key: telemetry.audit_logs.streams.datadog
  data:
  - stream_name: datadog
    api_key: "super-secret-key"
`))
		So(err, ShouldBeNil)

		out, err := NewSecretConfig(cfg, []config.SecretKey{config.TelemetryAuditLogStreamDatadogCredentialsKey}, time.Now())
		So(err, ShouldBeNil)
		So(out.TelemetryAuditLogStreamSecrets, ShouldResemble, &TelemetryAuditLogStreamSecrets{
			Datadog: []TelemetryAuditLogStreamDatadogSecret{{StreamName: "datadog"}},
		})

		b, err := json.Marshal(out)
		So(err, ShouldBeNil)
		So(string(b), ShouldNotContainSubstring, "super-secret-key")
	})

	Convey("NewPortalFeatureConfig passes telemetry through", t, func() {
		disabled := true
		fc := &config.FeatureConfig{Telemetry: &config.TelemetryFeatureConfig{
			AuditLogs: &config.TelemetryAuditLogsFeatureConfig{
				Streaming: &config.TelemetryAuditLogsStreamingFeatureConfig{Disabled: &disabled},
			},
		}}
		So(*NewPortalFeatureConfig(fc).Telemetry.AuditLogs.Streaming.Disabled, ShouldBeTrue)
	})

	Convey("empty keepStreamNames survives the mutation's JSON round trip", t, func() {
		// graphql-go coerces a list input [] into []interface{}{}.
		input := map[string]interface{}{
			"telemetryAuditLogStreamSecrets": map[string]interface{}{
				"action": "cleanup",
				"cleanupData": map[string]interface{}{
					"keepStreamNames": []interface{}{},
				},
			},
		}
		b, err := json.Marshal(input)
		So(err, ShouldBeNil)

		var instructions config.SecretConfigUpdateInstruction
		So(json.Unmarshal(b, &instructions), ShouldBeNil)
		So(instructions.TelemetryAuditLogStreamSecretsUpdateInstruction.CleanupData.KeepStreamNames, ShouldNotBeNil)
	})
}
