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

func TestSentryHandler(t *testing.T) {
	Convey("SentryHandler", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(
			NewSkipLoggingMiddleware(),
		).Handler(&SentryHandler{
			Next: NewHandlerForTesting(&w),
		}))

		Convey("should handle normal log records", func() {
			logger.Info("test message")
			So(w.String(), ShouldEqual, "level=INFO msg=\"test message\"\n")
		})

		Convey("should skip logging when error should be skipped", func() {
			logger.Error("test message", Err(context.Canceled))
			So(w.String(), ShouldEqual, "")
		})

		Convey("should handle regular errors that should not be skipped", func() {
			regularErr := fmt.Errorf("regular error")
			logger.Error("test message", Err(regularErr))
			So(w.String(), ShouldEqual, "level=ERROR msg=\"test message\" error=\"regular error\"\n")
		})

		Convey("should handle WithErr logger", func() {
			regularErr := fmt.Errorf("regular error")
			errLogger := WithErr(logger, regularErr)

			errLogger.Error("test message")
			So(w.String(), ShouldEqual, "level=ERROR msg=\"test message\" error=\"regular error\"\n")
		})

		Convey("should skip logging with WithErr logger when error should be skipped", func() {
			errLogger := WithErr(logger, context.Canceled)

			errLogger.Error("test message")
			So(w.String(), ShouldEqual, "")
		})
	})
}
