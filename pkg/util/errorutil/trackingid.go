package errorutil

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

func FormatTrackingID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s-%s", sc.TraceID().String(), sc.SpanID().String())
}

type TrackingIDGetter interface {
	error
	TrackingID() string
}

type errorTrackingID struct {
	inner      error
	trackingID string
}

func WithTrackingID(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	trackingID := FormatTrackingID(ctx)
	if trackingID == "" {
		return err
	}

	return &errorTrackingID{err, trackingID}
}

func (e *errorTrackingID) Error() string      { return e.inner.Error() }
func (e *errorTrackingID) Unwrap() error      { return e.inner }
func (e *errorTrackingID) TrackingID() string { return e.trackingID }

func CollectTrackingID(err error) string {
	var id string
	Unwrap(err, func(err error) {
		if id != "" {
			return
		}
		if getter, ok := err.(TrackingIDGetter); ok {
			id = getter.TrackingID()
		}
	})

	return id
}
