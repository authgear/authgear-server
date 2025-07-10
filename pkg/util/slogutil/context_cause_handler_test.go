package slogutil

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewContextCauseMiddleware(t *testing.T) {
	Convey("NewContextCauseMiddleware", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(NewContextCauseMiddleware()).Handler(slog.NewTextHandler(&w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		ctx := context.Background()
		_ = ctx

		Convey("In slog, context is never nil", func() {
			logger.Info("testing")

			So(strings.Contains(w.String(), "context_cause=<context-is-nil>"), ShouldBeFalse)
			So(strings.Contains(w.String(), "context_cause=<context-err-is-nil>"), ShouldBeTrue)
		})

		Convey("context canceled without cause", func() {
			ctx, cancel := context.WithCancel(ctx)
			cancel()

			logger.InfoContext(ctx, "testing")

			// This test observes as a documentation of the stdlib behavior.
			// When it is canceled without cause, the cause is context.Canceled itself.
			So(strings.Contains(w.String(), `context_cause="context canceled"`), ShouldBeTrue)
		})

		Convey("context canceled with cause", func() {
			ctx, cancel := context.WithCancelCause(ctx)
			cancel(fmt.Errorf("the cause"))

			logger.InfoContext(ctx, "testing")
			So(strings.Contains(w.String(), `context_cause="the cause"`), ShouldBeTrue)
		})

		// Cannot test WithDeadline and and WithTimeout because we have no access to the clock
		// used by the package context.
	})
}
