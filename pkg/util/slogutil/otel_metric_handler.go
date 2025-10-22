package slogutil

import (
	"context"
	"log/slog"

	"github.com/jba/slog/withsupport"
	"go.opentelemetry.io/otel"
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

type OtelMetricHandlerTrackFuncType func(ctx context.Context, errorName MetricErrorName, err error)

// OtelMetricHandlerTrackFunc is the real implementation.
var OtelMetricHandlerTrackFunc OtelMetricHandlerTrackFuncType = func(ctx context.Context, errorName MetricErrorName, err error) {
	otelutil.IntCounterAddOne(
		ctx,
		CounterErrorCount,
		MetricOptionsForError(err)...,
	)
}

// OtelMetricHandler is a custom handler written according to the guideline of https://golang.org/s/slog-handler-guide
// In particular, the use of github.com/jba/slog/withsupport is recommended by the guide.
type OtelMetricHandler struct {
	trackFunc    OtelMetricHandlerTrackFuncType
	groupOrAttrs *withsupport.GroupOrAttrs
}

var _ slog.Handler = (*OtelMetricHandler)(nil)

func NewOtelMetricHandler() *OtelMetricHandler {
	return &OtelMetricHandler{
		trackFunc: OtelMetricHandlerTrackFunc,
	}
}

func (h *OtelMetricHandler) Enabled(context.Context, slog.Level) bool {
	// It is always enabled.
	return true
}

func (h *OtelMetricHandler) Handle(ctx context.Context, record slog.Record) error {
	// This processing order is recommended by the guide.
	// We should process the groupOrAttrs before record.

	h.groupOrAttrs.Apply(func(groups []string, attr slog.Attr) {
		if attr.Key == AttrKeyError {
			if err, ok := attr.Value.Any().(error); ok {
				errorName, ok := GetMetricErrorName(err)
				if ok {
					h.trackFunc(ctx, errorName, err)
				}
			}
		}
	})

	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == AttrKeyError {
			if err, ok := attr.Value.Any().(error); ok {
				errorName, ok := GetMetricErrorName(err)
				if ok {
					h.trackFunc(ctx, errorName, err)
				}
			}
		}
		return true
	})

	return nil
}

func (h *OtelMetricHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OtelMetricHandler{
		trackFunc:    h.trackFunc,
		groupOrAttrs: h.groupOrAttrs.WithAttrs(attrs),
	}
}

func (h *OtelMetricHandler) WithGroup(name string) slog.Handler {
	return &OtelMetricHandler{
		trackFunc:    h.trackFunc,
		groupOrAttrs: h.groupOrAttrs.WithGroup(name),
	}
}
