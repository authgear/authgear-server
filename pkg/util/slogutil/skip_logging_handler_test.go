package slogutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/lib/pq"
	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type mockLoggingSkippable struct {
	skip bool
}

func (m mockLoggingSkippable) SkipLogging() bool {
	return m.skip
}

func (m mockLoggingSkippable) Error() string {
	return "mock error"
}

func TestNewSkipLoggingMiddleware(t *testing.T) {
	Convey("NewSkipLoggingMiddleware", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(NewSkipLoggingMiddleware()).Handler(NewHandlerForTesting(slog.LevelInfo, &w)))

		Convey("respect wrapped handler Enabled()", func() {
			logger.Debug("testing")
			So(w.String(), ShouldEqual, "")
		})

		Convey("normal logging should pass through", func() {
			logger.Info("test message")
			So(w.String(), ShouldEqual, "level=INFO msg=\"test message\"\n")
		})

		testCases := []struct {
			name     string
			err      error
			expected string
		}{
			{
				name:     "context.Canceled should be skipped",
				err:      context.Canceled,
				expected: "level=ERROR msg=\"test message\" error=\"context canceled\" __authgear_skip_logging=true\n",
			},
			{
				name:     "context.DeadlineExceeded should be skipped",
				err:      context.DeadlineExceeded,
				expected: "level=ERROR msg=\"test message\" error=\"context deadline exceeded\" __authgear_skip_logging=true\n",
			},
			{
				name:     "http.ErrAbortHandler should be skipped",
				err:      http.ErrAbortHandler,
				expected: "level=ERROR msg=\"test message\" error=\"net/http: abort Handler\" __authgear_skip_logging=true\n",
			},
			{
				name:     "http.MaxBytesError should be skipped",
				err:      &http.MaxBytesError{Limit: 100},
				expected: "level=ERROR msg=\"test message\" error=\"http: request body too large\" __authgear_skip_logging=true\n",
			},
			{
				name:     "json.SyntaxError should be skipped",
				err:      &json.SyntaxError{},
				expected: "level=ERROR msg=\"test message\" error=\"\" __authgear_skip_logging=true\n",
			},
			{
				name:     "pq.Error with query_canceled should be skipped",
				err:      &pq.Error{Code: "57014"},
				expected: "level=ERROR msg=\"test message\" error=\"pq: \" __authgear_skip_logging=true\n",
			},
			{
				name:     "sql.ErrTxDone should be skipped",
				err:      sql.ErrTxDone,
				expected: "level=ERROR msg=\"test message\" error=\"sql: transaction has already been committed or rolled back\" __authgear_skip_logging=true\n",
			},
			{
				name:     "LoggingSkippable with skip=true should be skipped",
				err:      mockLoggingSkippable{skip: true},
				expected: "level=ERROR msg=\"test message\" error=\"mock error\" __authgear_skip_logging=true\n",
			},
			{
				name:     "LoggingSkippable with skip=false should not be skipped",
				err:      mockLoggingSkippable{skip: false},
				expected: "level=ERROR msg=\"test message\" error=\"mock error\"\n",
			},
			{
				name:     "regular error should not be skipped",
				err:      fmt.Errorf("regular error"),
				expected: "level=ERROR msg=\"test message\" error=\"regular error\"\n",
			},
		}

		Convey("error logging should be skipped when Err() is used", func() {
			for _, tc := range testCases {
				Convey(tc.name, func() {
					logger.Error("test message", Err(tc.err))
					So(w.String(), ShouldEqual, tc.expected)
				})
			}
		})

		Convey("error logging should be skipped when WithErr() is used", func() {
			for _, tc := range testCases {
				Convey(tc.name, func() {
					WithErr(logger, tc.err).Error("test message")
					So(w.String(), ShouldEqual, tc.expected)
				})
			}
		})

		Convey("error logging should be skipped when WithSkipLogging is used", func() {
			logger.With(SkipLogging()).Error("test message")
			// It may seem incorrect to have 2 __authgear_skip_logging here.
			// This is the result of the trivial implementation of WithAttrs.
			// In WithAttrs, we forward the call to both groupOrAttrs and the underlying handler.
			// The underlying handler (TextHandler in this case) implements WithAttrs by including the attrs when it prints.
			// On the other hand, we always include __authgear_skip_logging in the record's attrs so that IsLoggingSkipped(record) works.
			So(w.String(), ShouldEqual, "level=ERROR msg=\"test message\" __authgear_skip_logging=true __authgear_skip_logging=true\n")
		})

		Convey("ForceLogging should override skip behavior", func() {
			forcedErr := errorutil.ForceLogging(context.Canceled)
			logger.Error("test message", Err(forcedErr))
			So(w.String(), ShouldContainSubstring, "level=ERROR")
			So(w.String(), ShouldContainSubstring, "msg=\"test message\"")
		})
	})
}

func TestIgnoreError(t *testing.T) {
	Convey("IgnoreError", t, func() {
		Convey("should return false for ForceLogging errors", func() {
			err := errorutil.ForceLogging(context.Canceled)
			So(IgnoreError(err), ShouldBeFalse)
		})

		Convey("should return true for context.Canceled", func() {
			So(IgnoreError(context.Canceled), ShouldBeTrue)
		})

		Convey("should return true for context.DeadlineExceeded", func() {
			So(IgnoreError(context.DeadlineExceeded), ShouldBeTrue)
		})

		Convey("should return true for http.ErrAbortHandler", func() {
			So(IgnoreError(http.ErrAbortHandler), ShouldBeTrue)
		})

		Convey("should return true for http.MaxBytesError", func() {
			err := &http.MaxBytesError{Limit: 100}
			So(IgnoreError(err), ShouldBeTrue)
		})

		Convey("should return true for json.SyntaxError", func() {
			err := &json.SyntaxError{}
			So(IgnoreError(err), ShouldBeTrue)
		})

		Convey("should return true for pq.Error with query_canceled code", func() {
			err := &pq.Error{Code: "57014"}
			So(IgnoreError(err), ShouldBeTrue)
		})

		Convey("should return false for pq.Error with other codes", func() {
			err := &pq.Error{Code: "23505"}
			So(IgnoreError(err), ShouldBeFalse)
		})

		Convey("should return false for syscall.EPIPE", func() {
			err := syscall.EPIPE
			So(IgnoreError(err), ShouldBeFalse)
		})

		Convey("should return false for syscall.ECONNREFUSED", func() {
			err := syscall.ECONNREFUSED
			So(IgnoreError(err), ShouldBeFalse)
		})

		Convey("should return false for syscall.ECONNRESET", func() {
			err := syscall.ECONNRESET
			So(IgnoreError(err), ShouldBeFalse)
		})

		Convey("should return true for os.ErrDeadlineExceeded", func() {
			err := os.ErrDeadlineExceeded
			So(IgnoreError(err), ShouldBeTrue)
		})

		Convey("should return true for sql.ErrTxDone", func() {
			So(IgnoreError(sql.ErrTxDone), ShouldBeTrue)
		})

		Convey("should return true for *net.OpError", func() {
			So(IgnoreError(&net.OpError{
				Op:  "write",
				Net: "tcp",
				Err: syscall.ECONNRESET,
			}), ShouldBeTrue)
		})

		Convey("should handle LoggingSkippable interface", func() {
			skippable := mockLoggingSkippable{skip: true}
			So(IgnoreError(skippable), ShouldBeTrue)

			notSkippable := mockLoggingSkippable{skip: false}
			So(IgnoreError(notSkippable), ShouldBeFalse)
		})

		Convey("should return false for regular errors", func() {
			err := fmt.Errorf("regular error")
			So(IgnoreError(err), ShouldBeFalse)
		})

		Convey("should handle wrapped errors", func() {
			wrappedCanceled := fmt.Errorf("wrapped: %w", context.Canceled)
			So(IgnoreError(wrappedCanceled), ShouldBeTrue)
		})
	})
}

func TestIsLoggingSkipped(t *testing.T) {
	timeZero := time.Time{}
	Convey("IsLoggingSkipped", t, func() {
		Convey("should return false for record without skip attribute", func() {
			record := slog.NewRecord(timeZero, slog.LevelInfo, "test message", 0)
			So(IsLoggingSkipped(record), ShouldBeFalse)
		})

		Convey("should return true for record with skip attribute set to true", func() {
			record := slog.NewRecord(timeZero, slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.Bool(AttrKeySkipLogging, true))
			So(IsLoggingSkipped(record), ShouldBeTrue)
		})

		Convey("should return false for record with skip attribute set to false", func() {
			record := slog.NewRecord(timeZero, slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.Bool(AttrKeySkipLogging, false))
			So(IsLoggingSkipped(record), ShouldBeFalse)
		})

		Convey("should return false for record with skip attribute of wrong type", func() {
			record := slog.NewRecord(timeZero, slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.String(AttrKeySkipLogging, "true"))
			So(IsLoggingSkipped(record), ShouldBeFalse)
		})

		Convey("should return false for record with skip attribute as int", func() {
			record := slog.NewRecord(timeZero, slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.Int(AttrKeySkipLogging, 1))
			So(IsLoggingSkipped(record), ShouldBeFalse)
		})

		Convey("should handle record with multiple attributes", func() {
			record := slog.NewRecord(timeZero, slog.LevelInfo, "test message", 0)
			record.AddAttrs(
				slog.String("other", "value"),
				slog.Bool(AttrKeySkipLogging, true),
				slog.Int("number", 42),
			)
			So(IsLoggingSkipped(record), ShouldBeTrue)
		})
	})
}
