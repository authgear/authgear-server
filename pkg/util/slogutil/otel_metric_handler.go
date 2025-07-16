package slogutil

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

var meter = otel.Meter("github.com/authgear/authgear-server/pkg/util/slogutil")

// CounterErrorCount is a metric to track error rate.
// It has these labels.
// - error_name
var CounterErrorCount = otelutil.MustInt64Counter(
	meter,
	"authgear.error.count",
	metric.WithDescription("The number of error encountered"),
	metric.WithUnit("{error}"),
)

type metricOptionAttributeKeyValue struct {
	attribute.KeyValue
}

func (o metricOptionAttributeKeyValue) ToOtelMetricOption() metric.MeasurementOption {
	return metric.WithAttributes(o.KeyValue)
}

func WithErrorName(errorName MetricErrorName) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{attribute.Key("error_name").String(string(errorName))}
}

type OtelMetricHandler struct{}

func (h OtelMetricHandler) Enabled(context.Context, slog.Level) bool {
	// It is always enabled.
	return true
}

func (h OtelMetricHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == AttrKeyError {
			if err, ok := attr.Value.Any().(error); ok {
				errorName, ok := GetMetricErrorName(err)
				if ok {
					otelutil.IntCounterAddOne(
						ctx,
						CounterErrorCount,
						WithErrorName(errorName),
					)
				}
			}
		}
		return true
	})

	return nil
}

func (h OtelMetricHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// This is a sink. It does not handle attrs.
	return h
}

func (h OtelMetricHandler) WithGroup(name string) slog.Handler {
	// This is a sink. It does not handle group.
	return h
}

var _ slog.Handler = OtelMetricHandler{}
