package otelutil

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
)

func MustInt64Counter(meter metric.Meter, name string, options ...metric.Int64CounterOption) metric.Int64Counter {
	counter, err := meter.Int64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func MustFloat64Histogram(meter metric.Meter, name string, options ...metric.Float64HistogramOption) metric.Float64Histogram {
	histogram, err := meter.Float64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return histogram
}

// IntCounter is metric.Int64Counter or metric.Int64UpDownCounter
type IntCounter interface {
	Add(ctx context.Context, incr int64, options ...metric.AddOption)
}

type MetricOption interface {
	ToOtelMetricOption() metric.MeasurementOption
}

// IntCounterAddOne prepares necessary attributes and calls Add with incr=1.
// It is intentionally that this does not accept metric.AddOption.
// If this accepts metric.AddOption, then you can pass in arbitrary metric.WithAttributes.
// Those attributes MAY NOT be the attributes defined in this package, and could contain
// unexpected end user data.
func IntCounterAddOne(ctx context.Context, counter IntCounter, inOptions ...MetricOption) {
	var finalOptions []metric.AddOption

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	labelerAttrs := labeler.Get()
	for _, labelerAttr := range labelerAttrs {
		finalOptions = append(finalOptions, metric.WithAttributes(labelerAttr))
	}

	for _, o := range inOptions {
		finalOptions = append(finalOptions, o.ToOtelMetricOption())
	}

	counter.Add(ctx, 1, finalOptions...)
}

func Float64HistogramRecord(ctx context.Context, histogram metric.Float64Histogram, val float64, inOptions ...MetricOption) {
	var finalOptions []metric.RecordOption

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	labelerAttrs := labeler.Get()
	for _, labelerAttr := range labelerAttrs {
		finalOptions = append(finalOptions, metric.WithAttributes(labelerAttr))
	}

	for _, o := range inOptions {
		finalOptions = append(finalOptions, o.ToOtelMetricOption())
	}

	histogram.Record(ctx, val, finalOptions...)
}
