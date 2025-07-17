package slogutil

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewContextCauseMiddleware(t *testing.T) {
	Convey("NewContextCauseMiddleware", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(NewContextCauseMiddleware()).Handler(NewHandlerForTesting(slog.LevelInfo, &w)))

		ctx := context.Background()
		_ = ctx

		Convey("respect wrapped handler Enabled()", func() {
			logger.Debug("testing")

			So(w.String(), ShouldEqual, "")
		})

		Convey("ctx is never nil when logging is done with the Logger frontend", func() {
			logger.Info("testing")

			So(w.String(), ShouldEqual, "level=INFO msg=testing\n")
		})

		Convey("context canceled without cause", func() {
			ctx, cancel := context.WithCancel(ctx)
			cancel()

			logger.InfoContext(ctx, "testing")

			// This test observes as a documentation of the stdlib behavior.
			// When it is canceled without cause, the cause is context.Canceled itself.
			So(w.String(), ShouldEqual, "level=INFO msg=testing context_cause=\"context canceled\"\n")
		})

		Convey("context canceled with cause", func() {
			ctx, cancel := context.WithCancelCause(ctx)
			cancel(fmt.Errorf("the cause"))

			logger.InfoContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=INFO msg=testing context_cause=\"the cause\"\n")
		})

		Convey("context timed out without cause", func() {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
			defer cancel()

			time.Sleep(1 * time.Microsecond)
			logger.InfoContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=INFO msg=testing context_cause=\"context deadline exceeded\"\n")
		})

		Convey("context timed out with cause", func() {
			ctx, cancel := context.WithTimeoutCause(ctx, 1*time.Nanosecond, fmt.Errorf("the cause"))
			defer cancel()

			time.Sleep(1 * time.Microsecond)
			logger.InfoContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=INFO msg=testing context_cause=\"the cause\"\n")
		})

		Convey("context deadline without cause", func() {
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Nanosecond))
			defer cancel()

			time.Sleep(1 * time.Microsecond)
			logger.InfoContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=INFO msg=testing context_cause=\"context deadline exceeded\"\n")
		})

		Convey("context deadline with cause", func() {
			ctx, cancel := context.WithDeadlineCause(ctx, time.Now().Add(1*time.Nanosecond), fmt.Errorf("the cause"))
			defer cancel()

			time.Sleep(1 * time.Microsecond)
			logger.InfoContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=INFO msg=testing context_cause=\"the cause\"\n")
		})

		Convey("does not duplicate attrs", func() {
			logger.Info("testing", slog.String("foobar", "42"))

			So(w.String(), ShouldEqual, "level=INFO msg=testing foobar=42\n")
		})
	})
}
