package slogutil

import (
	"context"
	"log/slog"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewOtelMetricHandler(t *testing.T) {
	Convey("NewOtelMetricHandler", t, func() {
		handler := NewOtelMetricHandler()

		Convey("should create handler with default track function", func() {
			So(handler, ShouldNotBeNil)
			So(handler.trackFunc, ShouldNotBeNil)
			So(handler.groupOrAttrs, ShouldBeNil)
		})
	})
}

func TestOtelMetricHandler_Integration(t *testing.T) {
	Convey("OtelMetricHandler Integration", t, func() {
		var trackedErrors []MetricErrorName
		mockTrackFunc := func(ctx context.Context, errorName MetricErrorName) {
			trackedErrors = append(trackedErrors, errorName)
		}

		logger := slog.New(&OtelMetricHandler{
			trackFunc: mockTrackFunc,
		})

		Convey("should work with chained WithAttrs and WithGroup", func() {
			ctx := context.Background()

			logger = logger.
				With(Err(context.Canceled)).
				WithGroup("group1").
				With(slog.String("key", "value"))

			logger.ErrorContext(ctx, "msg", Err(context.DeadlineExceeded))
			So(len(trackedErrors), ShouldEqual, 2)
			So(trackedErrors[0], ShouldEqual, MetricErrorNameContextCanceled)
			So(trackedErrors[1], ShouldEqual, MetricErrorNameContextDeadlineExceeded)
		})
	})
}
