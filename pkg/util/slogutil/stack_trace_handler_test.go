package slogutil

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewStackTraceMiddleware(t *testing.T) {
	Convey("NewHandleInlineMiddleware", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(NewStackTraceMiddleware()).Handler(NewHandlerForTesting(slog.LevelWarn, &w)))

		ctx := context.Background()

		Convey("respects wrapped handler Enabled()", func() {
			logger.DebugContext(ctx, "should not log this")

			So(w.String(), ShouldEqual, "")
		})

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

		Convey("attrs are not duplicated", func() {
			logger.ErrorContext(ctx, "testing", slog.String("foobar", "42"))

			So(strings.Contains(w.String(), "stack="), ShouldBeTrue)
			So(strings.Contains(w.String(), "foobar=42"), ShouldBeTrue)

			re := regexp.MustCompile("foobar=42")
			matches := re.FindAllString(w.String(), -1)
			So(len(matches), ShouldEqual, 1)
		})

		Convey("SkipStackTrace with attr", func() {
			logger.ErrorContext(ctx, "testing", SkipStackTrace())

			So(w.String(), ShouldEqual, "level=ERROR msg=testing __authgear_skip_stacktrace=true\n")
		})

		Convey("SkipStackTrace with logger", func() {
			logger = logger.With(SkipStackTrace())
			logger.ErrorContext(ctx, "testing")

			So(w.String(), ShouldEqual, "level=ERROR msg=testing __authgear_skip_stacktrace=true\n")
		})
	})
}
