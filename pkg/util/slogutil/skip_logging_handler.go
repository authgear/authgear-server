package slogutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lib/pq"
	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

// LoggingSkippable is an interface to be implemented by error.
type LoggingSkippable interface{ SkipLogging() bool }

const AttrKeySkipLogging = "__authgear_skip_logging"

type SkipLoggingHandler struct {
	Next slog.Handler
}

var _ slog.Handler = SkipLoggingHandler{}

func NewSkipLoggingMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return SkipLoggingHandler{
			Next: next,
		}
	}
}

func (s SkipLoggingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// We want this handler to always run.
	return true
}

func (s SkipLoggingHandler) Handle(ctx context.Context, record slog.Record) error {
	shouldSkip := false

	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == AttrKeyError {
			if err, ok := attr.Value.Any().(error); ok {
				if IgnoreError(err) {
					shouldSkip = true
				}
			}
			// We have found the key, we can stop the iteration.
			return false
		}
		return true
	})

	// We always call the next handler.
	// The way we skip handler is to add an attribute for downstream handler to read.
	if shouldSkip {
		record = record.Clone()
		record.AddAttrs(slog.Bool(AttrKeySkipLogging, true))
	}

	return s.Next.Handle(ctx, record)
}

func (s SkipLoggingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return SkipLoggingHandler{
		Next: s.Next.WithAttrs(attrs),
	}
}

func (s SkipLoggingHandler) WithGroup(name string) slog.Handler {
	return SkipLoggingHandler{
		Next: s.Next.WithGroup(name),
	}
}

func IgnoreError(err error) (ignore bool) {
	// IsForceLogging overrides everything.
	if errorutil.IsForceLogging(err) {
		ignore = false
		return
	}

	// In most case, when the HTTP request was canceled,
	// the error would be context.Canceled.
	// However, by observation, other errors may occur too.
	// We consider all of them as request canceled and do not log them.
	if errors.Is(err, context.Canceled) {
		ignore = true
	}

	// DeadlineExceeded occurs when we shutdown the daemon but encounter a timeout.
	if errors.Is(err, context.DeadlineExceeded) {
		ignore = true
	}

	// ErrAbortHandler is a sentinel panic value to abort a handler.
	// net/http does NOT log this error, so do we.
	if errors.Is(err, http.ErrAbortHandler) {
		ignore = true
	}

	// http.MaxBytesReader will *http.MaxBytesError when the body is too large.
	// We do not want to log this.
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		ignore = true
	}

	// json.Unmarshal returns a SyntaxError if the JSON can't be parsed.
	// https://pkg.go.dev/encoding/json#SyntaxError
	var jsonSyntaxError *json.SyntaxError
	if errors.As(err, &jsonSyntaxError) {
		ignore = true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		// https://www.postgresql.org/docs/13/errcodes-appendix.html
		// 57014 is query_canceled
		if pqError.Code == "57014" {
			ignore = true
		}
	}

	if errors.Is(err, sql.ErrTxDone) {
		// The sql package will rollback the transaction when the context is canceled.
		// https://pkg.go.dev/database/sql/driver#ConnBeginTx
		// So when we call Rollback() again, this error will be returned.
		ignore = true
	}

	var skippable LoggingSkippable
	if errors.As(err, &skippable) {
		if skippable.SkipLogging() {
			ignore = true
		}
	}

	return
}

func IsLoggingSkipped(record slog.Record) bool {
	skipped := false
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == AttrKeySkipLogging {
			if attr.Value.Kind() == slog.KindBool {
				skipped = attr.Value.Bool()
				return false
			}
		}
		return true
	})
	return skipped
}
