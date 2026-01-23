package otelutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	instrumentationruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/authgear/authgear-server/pkg/version"
)

// envvar_OTEL_METRICS_EXPORTER is documented at
// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#exporter-selection
const envvar_OTEL_METRICS_EXPORTER = "OTEL_METRICS_EXPORTER"

// envvar_OTEL_PROPAGATORS is documented at
// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration
const envvar_OTEL_PROPAGATORS = "OTEL_PROPAGATORS"

// SetupOTelSDKGlobally sets up the global propagator and the global meter provider.
// Setting these globally allows us to define metric globally.
// Additionally, it returns a context that MUST BE used as the background context.
// The returned context contains a *sdkresource.Resource.
// The returned context contains a *otelhttp.Labeler.
func SetupOTelSDKGlobally(ctx context.Context) (outCtx context.Context, shutdown func(context.Context) error, err error) {
	// Force producing the deprecated metrics.
	// See https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime@v0.62.0#pkg-overview
	// We have to do this because the number of GC count is only available in the deprecated metrics.
	os.Setenv("OTEL_GO_X_DEPRECATED_RUNTIME_METRICS", "true")

	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Ensure shutdown is called when we encounter an error.
	defer func() {
		if err != nil {
			err = errors.Join(err, shutdown(ctx))
		}
	}()

	// Set up resource.
	res, err := newResource(ctx)
	if err != nil {
		return
	}
	outCtx = ctx

	// Set up propagator.
	propagator, err := newPropagator()
	if err != nil {
		return
	}
	// Set the global propagator.
	otel.SetTextMapPropagator(propagator)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	// set the global meter provider.
	otel.SetMeterProvider(meterProvider)

	// Set up trace provider.
	traceProvider, err := newTracerProvider()
	if err != nil {
		return
	}
	otel.SetTracerProvider(traceProvider)

	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)

	// TODO: Support export logs with http
	logExporter, err := stdoutlog.New()
	if err != nil {
		return
	}

	logProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	outCtx = WithOTelLoggerProvider(outCtx, logProvider)

	shutdownFuncs = append(shutdownFuncs, logProvider.Shutdown)

	// Start go runtime metrics collection.
	// Refer to https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime@v0.62.0#pkg-overview
	// for a list of metrics it collects.
	err = instrumentationruntime.Start(
		instrumentationruntime.WithMinimumReadMemStatsInterval(instrumentationruntime.DefaultMinimumReadMemStatsInterval),
	)
	if err != nil {
		return
	}

	return
}

func newResource(ctx context.Context) (*sdkresource.Resource, error) {
	return sdkresource.New(
		ctx,

		// Include the git-hash
		sdkresource.WithAttributes(
			semconv.ServiceVersionKey.String(version.Version),
		),

		// Information about the otel SDK itself.
		sdkresource.WithTelemetrySDK(),

		// OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME
		sdkresource.WithFromEnv(),

		// The following is WithProcess, except that WithProcessCommandArgs is excluded.
		// The arguments MAY contain sensitive information (such as database password) that
		// we DO NOT want to include.
		sdkresource.WithProcessPID(),
		sdkresource.WithProcessExecutableName(),
		sdkresource.WithProcessExecutablePath(),
		sdkresource.WithProcessOwner(),
		sdkresource.WithProcessRuntimeName(),
		sdkresource.WithProcessRuntimeVersion(),
		sdkresource.WithProcessRuntimeDescription(),

		// Information about the OS.
		sdkresource.WithOS(),

		// Information about container, if it is run as a container.
		sdkresource.WithContainer(),

		// os.Hostname
		sdkresource.WithHost(),

		// /etc/machine-id or /var/lib/dbus/machine-id
		// Since it could fail, we do not include it now.
		// sdkresource.WithHostID(),
	)
}

func newPropagator() (out propagation.TextMapPropagator, err error) {
	// The specification says the default value of OTEL_PROPAGATORS is "tracecontext,baggage"
	// And that is a sane default.

	OTEL_PROPAGATORS := strings.TrimSpace(os.Getenv(envvar_OTEL_PROPAGATORS))

	// Handle default value.
	if OTEL_PROPAGATORS == "" {
		OTEL_PROPAGATORS = "tracecontext,baggage"
	}

	// Handle "none"
	if OTEL_PROPAGATORS == "none" {
		// This is the official way to construct a no-op propagator.
		// See https://github.com/open-telemetry/opentelemetry-go/blob/v1.32.0/internal/global/propagator.go#L29
		out = propagation.NewCompositeTextMapPropagator()
		return
	}

	var propagators []propagation.TextMapPropagator
	parts := strings.Split(OTEL_PROPAGATORS, ",")
	for _, part := range parts {
		switch part {
		case "tracecontext":
			propagators = append(propagators, propagation.TraceContext{})
		case "baggage":
			propagators = append(propagators, propagation.Baggage{})
		default:
			err = fmt.Errorf("unsupported value: %v=%v", envvar_OTEL_PROPAGATORS, OTEL_PROPAGATORS)
			return
		}
	}

	out = propagation.NewCompositeTextMapPropagator(propagators...)
	return
}

func newMeterProvider(ctx context.Context, res *sdkresource.Resource) (*sdkmetric.MeterProvider, error) {
	options := []sdkmetric.Option{
		sdkmetric.WithResource(res),
	}

	exporters, err := newMetricExportersFromEnv(ctx)
	if err != nil {
		return nil, err
	}
	for _, exporter := range exporters {
		// Use PeriodicReader because it supports configuration via environment variables.
		// See https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#periodic-exporting-metricreader
		reader := sdkmetric.NewPeriodicReader(
			exporter,
			// The example says that we can optionally set up a Producer to track the latency of goroutine.
			// See https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime@v0.62.0#example-package
			// The histogram is `go.schedule.duration`.
			// However, the histogram has no explicit boundary set,
			// so it can consume a lot of memory.
			// Therefore, this feature is commented out.
			// sdkmetric.WithProducer(instrumentationruntime.NewProducer()),
		)
		options = append(options, sdkmetric.WithReader(reader))
	}

	meterProvider := sdkmetric.NewMeterProvider(options...)
	return meterProvider, nil
}

func newMetricExportersFromEnv(ctx context.Context) (exporters []sdkmetric.Exporter, err error) {
	// The specification says the default value of OTEL_METRICS_EXPORTER is "otlp".
	// The documentation of the Go SDK says NewMeterProvider does not have any Reader.
	// Without any Reader, it does nothing.
	// See https://pkg.go.dev/go.opentelemetry.io/otel/sdk/metric#NewMeterProvider
	//
	// I think the behavior of the SDK is a more sane default.
	// The export of metrics should be OPT-IN, rather than OPT-OUT.
	// This makes Authgear backwards-compatible if OTEL_METRICS_EXPORTER is not set.
	OTEL_METRICS_EXPORTER := strings.TrimSpace(os.Getenv(envvar_OTEL_METRICS_EXPORTER))
	if OTEL_METRICS_EXPORTER == "" || OTEL_METRICS_EXPORTER == "none" {
		return nil, nil
	}

	// The spec says the implementation SHOULD support comma-separated list.
	parts := strings.Split(OTEL_METRICS_EXPORTER, ",")
	for _, part := range parts {
		switch part {
		case "otlp":
			exporter, err := otlpmetrichttp.New(ctx)
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		case "console":
			exporter, err := stdoutmetric.New()
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		default:
			err = fmt.Errorf("unsupported value: %v=%v", envvar_OTEL_METRICS_EXPORTER, OTEL_METRICS_EXPORTER)
			return
		}
	}

	return
}

func newTracerProvider() (*trace.TracerProvider, error) {
	// We do not collect traces at the moment.
	// Configure exporters here when we want to collect trace data.
	tracerProvider := trace.NewTracerProvider()
	return tracerProvider, nil
}
