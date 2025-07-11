package slogutil

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewStackTraceMiddleware(t *testing.T) {
	Convey("NewHandleInlineMiddleware", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(NewStackTraceMiddleware()).Handler(NewHandlerForTesting(&w)))

		ctx := context.Background()

		Convey("does not include stack trace when level < error", func() {
			logger.WarnContext(ctx, "testing")

			So(w.String(), ShouldEqual, "level=WARN msg=testing\n")
		})

		Convey("include stack trace when level >= error", func() {
			logger.ErrorContext(ctx, "testing")

			// The actual stack trace is unstable and too long to be included.
			So(strings.Contains(w.String(), "stack="), ShouldBeTrue)
		})

		Convey("WithAttrs are retained", func() {
			logger = logger.With("foobar", "42")

			logger.ErrorContext(ctx, "testing")

			So(strings.Contains(w.String(), "stack="), ShouldBeTrue)
			So(strings.Contains(w.String(), "foobar=42"), ShouldBeTrue)
		})

		Convey("WithGroup are retained", func() {
			logger = logger.WithGroup("group")

			logger.ErrorContext(ctx, "testing")

			So(strings.Contains(w.String(), "group.stack="), ShouldBeTrue)
		})
	})
}
