package otelutil

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestNewMetricExportersFromEnv(t *testing.T) {
	Convey("newMetricExportersFromEnv", t, func() {
		test := func(envValue string, expected []sdkmetric.Exporter) {
			t.Setenv("OTEL_METRICS_EXPORTER", envValue)
			ctx := context.Background()
			actual, err := newMetricExportersFromEnv(ctx)
			So(err, ShouldBeNil)
			So(len(actual), ShouldEqual, len(expected))
			for i := 0; i < len(actual); i += 1 {
				So(actual[i], ShouldHaveSameTypeAs, expected[i])
			}
		}

		// stdoutmetric does not export the concrete type of the exporter.
		// So we need to grab the type in this way.
		stdoutmetricExporter, err := stdoutmetric.New()
		So(err, ShouldBeNil)

		test("", nil)
		test("none", nil)
		test("otlp", []sdkmetric.Exporter{
			&otlpmetrichttp.Exporter{},
		})
		test("console", []sdkmetric.Exporter{
			stdoutmetricExporter,
		})
		test("otlp,console", []sdkmetric.Exporter{
			&otlpmetrichttp.Exporter{},
			stdoutmetricExporter,
		})
		test("console,otlp", []sdkmetric.Exporter{
			stdoutmetricExporter,
			&otlpmetrichttp.Exporter{},
		})
		test("    console,otlp     ", []sdkmetric.Exporter{
			stdoutmetricExporter,
			&otlpmetrichttp.Exporter{},
		})
	})
}
