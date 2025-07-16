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
}

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
	if record.Level >= slog.LevelError {
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
	return &StackTraceHandler{
		Next: s.Next.WithAttrs(attrs),
	}
}

func (s *StackTraceHandler) WithGroup(name string) slog.Handler {
	return &StackTraceHandler{
		Next: s.Next.WithGroup(name),
	}
}
