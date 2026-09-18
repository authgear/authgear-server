package auditlogstreaming

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
)

var ProducerDependencySet = wire.NewSet(
	NewProducer,
)

// NewProducer assumes telemetryConfig and featureConfig have already been
// through config.SetFieldDefaults (as every parsed config has, by the time
// wire injects it), so every level down to Streaming.Disabled is non-nil.
func NewProducer(
	redis *globalredis.Handle,
	appID config.AppID,
	telemetryConfig *config.TelemetryConfig,
	featureConfig *config.FeatureConfig,
) *Producer {
	return &Producer{
		AppID:    appID,
		Streams:  telemetryConfig.AuditLogs.Streams,
		Disabled: *featureConfig.Telemetry.AuditLogs.Streaming.Disabled,
		Redis:    redis,
	}
}
