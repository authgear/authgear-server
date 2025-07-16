package slogutil

import (
	"context"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
)

type ContextCauseHandler struct {
	Next slog.Handler
}

var _ slog.Handler = (*ContextCauseHandler)(nil)

func NewContextCauseMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &ContextCauseHandler{
			Next: next,
		}
	}
}

func (s *ContextCauseHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (s *ContextCauseHandler) Handle(ctx context.Context, record slog.Record) error {
	attrValue := ""

	if ctx == nil {
		attrValue = "<context-is-nil>"
	} else {
		ctxErr := ctx.Err()
		if ctxErr == nil {
			attrValue = "<context-err-is-nil>"
		} else {
			cause := context.Cause(ctx)
			if cause == nil {
				attrValue = "<context-cause-is-nil>"
			} else {
				attrValue = cause.Error()
			}
		}
	}

	record = record.Clone()
	record.AddAttrs(slog.Attr{
		Key:   "context_cause",
		Value: slog.StringValue(attrValue),
	})

	if s.Next.Enabled(ctx, record.Level) {
		return s.Next.Handle(ctx, record)
	}

	return nil
}

func (s *ContextCauseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextCauseHandler{
		Next: s.Next.WithAttrs(attrs),
	}
}

func (s *ContextCauseHandler) WithGroup(name string) slog.Handler {
	return &ContextCauseHandler{
		Next: s.Next.WithGroup(name),
	}
}
