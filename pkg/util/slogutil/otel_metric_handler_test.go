package slogutil

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/attribute"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
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
		var trackedErrorOptions [][]otelutil.MetricOption
		mockTrackFunc := func(ctx context.Context, errorName MetricErrorName, err error) {
			trackedErrors = append(trackedErrors, errorName)
			trackedErrorOptions = append(trackedErrorOptions, MetricOptionsForError(err))
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

		Convey("should track net.OpError with Op and Net attributes", func() {
			ctx := context.Background()
			opErr := &net.OpError{Op: "read", Net: "tcp", Err: errors.New("connection reset by peer")}
			logger.ErrorContext(ctx, "network error", Err(opErr))

			So(len(trackedErrors), ShouldEqual, 1)
			So(trackedErrors[0], ShouldEqual, MetricErrorNameNetOpError)

			So(len(trackedErrorOptions), ShouldEqual, 1)
			opts := trackedErrorOptions[0]
			So(opts, ShouldContain, MetricOptionAttributeKeyValue{attribute.Key("net.op").String("read")})
			So(opts, ShouldContain, MetricOptionAttributeKeyValue{attribute.Key("net.net").String("tcp")})
		})
	})
}
