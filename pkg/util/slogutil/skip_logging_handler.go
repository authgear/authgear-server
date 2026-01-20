package slogutil

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jba/slog/withsupport"
	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

// LoggingSkippable is an interface to be implemented by error.
type LoggingSkippable interface{ SkipLogging() bool }

// attrKeySkipLogging is used internally in the logs middleware for filtering specific errors only.
// It should not be used when defining logs.
const attrKeySkipLogging = "__authgear_skip_logging"

type SkipLoggingHandler struct {
	// See https://github.com/golang/example/blob/master/slog-handler-guide/README.md#the-withgroup-method
	groupOrAttrs *withsupport.GroupOrAttrs
	Next         slog.Handler
}

var _ slog.Handler = (*SkipLoggingHandler)(nil)

func NewSkipLoggingMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &SkipLoggingHandler{
			Next: next,
		}
	}
}

func (s *SkipLoggingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// We want this handler to always run.
	return true
}

func (s *SkipLoggingHandler) Handle(ctx context.Context, record slog.Record) error {
	shouldSkip := false

	visitAttr := func(attr slog.Attr) {
		if attr.Key == AttrKeyError {
			if err, ok := attr.Value.Any().(error); ok {
				if IgnoreError(err) {
					shouldSkip = true
				}
			}
		}
	}

	s.groupOrAttrs.Apply(func(groups []string, attr slog.Attr) {
		visitAttr(attr)
	})

	record.Attrs(func(attr slog.Attr) bool {
		visitAttr(attr)
		return true
	})

	// We always call the next handler.
	// The way we skip handler is to add an attribute for downstream handler to read.
	if shouldSkip {
		// Clone the record without attrs.
		clonedWithoutAttrs := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)

		// Loop through the original attrs and add them to cloned.
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key != attrKeySkipLogging {
				clonedWithoutAttrs.AddAttrs(attr)
			}
			return true
		})

		// Always write the skip_logging attr.
		clonedWithoutAttrs.AddAttrs(slog.Bool(attrKeySkipLogging, true))

		record = clonedWithoutAttrs
	}

	if s.Next.Enabled(ctx, record.Level) {
		return s.Next.Handle(ctx, record)
	}
	return nil
}

func (s *SkipLoggingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SkipLoggingHandler{
		Next:         s.Next.WithAttrs(attrs),
		groupOrAttrs: s.groupOrAttrs.WithAttrs(attrs),
	}
}

func (s *SkipLoggingHandler) WithGroup(name string) slog.Handler {
	return &SkipLoggingHandler{
		Next:         s.Next.WithGroup(name),
		groupOrAttrs: s.groupOrAttrs.WithGroup(name),
	}
}

func IgnoreError(err error) (ignore bool) {
	// IsForceLogging overrides everything.
	if errorutil.IsForceLogging(err) {
		ignore = false
		return
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

	// Ignore any errors that are tracked as metrics.
	_, ok := GetMetricErrorName(err)
	if ok {
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
		if attr.Key == attrKeySkipLogging {
			if attr.Value.Kind() == slog.KindBool {
				skipped = attr.Value.Bool()
				return false
			}
		}
		return true
	})
	return skipped
}
