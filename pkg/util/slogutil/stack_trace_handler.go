package slogutil

import (
	"context"
	"log/slog"
	"strings"

	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type StackTraceHandler struct {
	Next slog.Handler
	Skip bool
}

const AttrKeySkipStackTrace = "__authgear_skip_stacktrace"

var _ slog.Handler = (*StackTraceHandler)(nil)

func NewStackTraceMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &StackTraceHandler{
			Next: next,
		}
	}
}

func (s *StackTraceHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (s *StackTraceHandler) Handle(ctx context.Context, record slog.Record) error {
	if !s.Skip && record.Level >= slog.LevelError && !IsStackTraceSkipped(record) {
		record = record.Clone()
		record.AddAttrs(slog.Attr{
			Key:   "stack",
			Value: slog.StringValue(strings.Join(errorutil.Callers(10000), "\n")),
		})
	}

	if s.Next.Enabled(ctx, record.Level) {
		return s.Next.Handle(ctx, record)
	}
	return nil
}

func (s *StackTraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	for _, attr := range attrs {
		if attr.Key == AttrKeySkipStackTrace {
			if attr.Value.Kind() == slog.KindBool && attr.Value.Bool() {
				return &StackTraceHandler{
					Next: s.Next.WithAttrs(attrs),
					Skip: true,
				}
			}
		}
	}
	return &StackTraceHandler{
		Next: s.Next.WithAttrs(attrs),
		Skip: s.Skip,
	}
}

func (s *StackTraceHandler) WithGroup(name string) slog.Handler {
	return &StackTraceHandler{
		Next: s.Next.WithGroup(name),
		Skip: s.Skip,
	}
}

func SkipStackTrace() slog.Attr {
	return slog.Bool(AttrKeySkipStackTrace, true)
}

func IsStackTraceSkipped(record slog.Record) bool {
	skipped := false
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == AttrKeySkipStackTrace {
			if attr.Value.Kind() == slog.KindBool {
				skipped = attr.Value.Bool()
				return false
			}
		}
		return true
	})
	return skipped
}
