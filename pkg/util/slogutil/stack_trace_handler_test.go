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
		logger := slog.New(slogmulti.Pipe(NewStackTraceMiddleware()).Handler(slog.NewTextHandler(&w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		ctx := context.Background()

		Convey("does not include stack trace when level < error", func() {
			logger.WarnContext(ctx, "testing")

			So(strings.Contains(w.String(), "stack="), ShouldBeFalse)
		})

		Convey("include stack trace when level >= error", func() {
			logger.ErrorContext(ctx, "testing")

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
