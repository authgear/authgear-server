package slogutil

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

var meter = otel.Meter("github.com/authgear/authgear-server/pkg/util/slogutil")

// CounterContextCanceledCount is a temporary metric to debug context canceled issue.
// It has no labels.
var CounterContextCanceledCount = otelutil.MustInt64Counter(
	meter,
	"authgear.context_canceled.count",
	metric.WithDescription("The number of context canceled error encountered"),
	metric.WithUnit("{error}"),
)

type TraceContextCanceledHandler struct{}

func (h TraceContextCanceledHandler) Enabled(context.Context, slog.Level) bool {
	// It is always enabled.
	return true
}

func (h TraceContextCanceledHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == AttrKeyError {
			if err, ok := attr.Value.Any().(error); ok {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					otelutil.IntCounterAddOne(
						ctx,
						CounterContextCanceledCount,
					)
				}
			}
		}
		return true
	})

	return nil
}

func (h TraceContextCanceledHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// This is a sink. It does not handle attrs.
	return h
}

func (h TraceContextCanceledHandler) WithGroup(name string) slog.Handler {
	// This is a sink. It does not handle group.
	return h
}

var _ slog.Handler = TraceContextCanceledHandler{}
