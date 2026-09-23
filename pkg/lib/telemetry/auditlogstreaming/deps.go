package auditlogstreaming

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
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
	interval config.DurationString,
) *Producer {
	return &Producer{
		AppID:    appID,
		Streams:  telemetryConfig.AuditLogs.Streams,
		Disabled: *featureConfig.Telemetry.AuditLogs.Streaming.Disabled,
		Redis:    redis,
		Interval: interval,
	}
}

// SenderDependencySet is used by the background binary's own wireinject
// function to build a *SenderImpl per app inside a SenderFactory.
var SenderDependencySet = wire.NewSet(
	NewSenderImpl,
)

// NewSenderImpl builds a Sender for one app from its resolved config: the
// same shape a SenderFactory.MakeSender implementation needs.
func NewSenderImpl(
	appID string,
	httpConfig *config.HTTPConfig,
	httpFeatureConfig *config.HTTPFeatureConfig,
	telemetryConfig *config.TelemetryConfig,
	tlsMaterials *config.TelemetryAuditLogStreamTLSMaterials,
	datadogCredentials *config.TelemetryAuditLogStreamDatadogCredentials,
) *SenderImpl {
	return &SenderImpl{
		AppID:              appID,
		Hostname:           ResolveHostname(httpConfig.PublicOrigin),
		Streams:            telemetryConfig.AuditLogs.Streams,
		TLS:                tlsMaterials,
		DatadogCredentials: datadogCredentials,
		DatadogClientFactory: &SSRFSafeDatadogClientFactory{
			AllowNonPublicAddresses: httpFeatureConfig.IsInsecureFetchAddressAllowed(),
			AllowedHosts:            httpFeatureConfig.GetInsecureFetchAddressAllowedHosts(),
		},
	}
}

// ConsumerDependencySet is used by the background binary's own wireinject
// function to construct the *backgroundjob.Runner that delivers queued
// audit log entries.
var ConsumerDependencySet = wire.NewSet(
	NewRunnableFactory,
	NewRunner,
)

// RunnableDependencySet is used by this package's own newRunnable
// (wire.go) to build a fresh Runnable per tick, following the same
// pattern as pkg/lib/feature/accountdeletion.
var RunnableDependencySet = wire.NewSet(
	wire.Struct(new(Consumer), "*"),
	wire.Struct(new(Runnable), "*"),
	wire.Bind(new(backgroundjob.Runnable), new(*Runnable)),
)

func NewRunnableFactory(
	redis *globalredis.Handle,
	interval config.DurationString,
	appContextResolver AppContextResolver,
	senderFactory SenderFactory,
) backgroundjob.RunnableFactory {
	factory := func() backgroundjob.Runnable {
		return newRunnable(redis, interval, appContextResolver, senderFactory)
	}
	return factory
}

func NewRunner(ctx context.Context, runnableFactory backgroundjob.RunnableFactory, interval config.DurationString) *backgroundjob.Runner {
	return backgroundjob.NewRunner(ctx, runnableFactory, backgroundjob.WithAfterDuration(interval.Duration()))
}
