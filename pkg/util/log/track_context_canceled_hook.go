package log

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

var meter = otel.Meter("github.com/authgear/authgear-server/pkg/util/log")

// CounterContextCanceledCount is a temporary metric to debug context canceled issue.
// It has no labels.
var CounterContextCanceledCount = otelutil.MustInt64Counter(
	meter,
	"authgear.context_canceled.count",
	metric.WithDescription("The number of context canceled error encountered"),
	metric.WithUnit("{error}"),
)

type TrackContextCanceledHook struct{}

var _ logrus.Hook = (*TrackContextCanceledHook)(nil)

func (h *TrackContextCanceledHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TrackContextCanceledHook) Fire(entry *logrus.Entry) error {
	maybeCtx := entry.Context
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		if maybeCtx != nil {
			ctx := maybeCtx
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				otelutil.IntCounterAddOne(
					ctx,
					CounterContextCanceledCount,
				)
			}
		}
	}
	return nil
}
