package slogutil

import (
	"context"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
)

type ContextCauseHandler struct {
	Next slog.Handler
}

var _ slog.Handler = ContextCauseHandler{}

func NewContextCauseMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return ContextCauseHandler{
			Next: next,
		}
	}
}

func (s ContextCauseHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (s ContextCauseHandler) Handle(ctx context.Context, record slog.Record) error {
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

	attrs := []slog.Attr{
		{
			Key:   "context_cause",
			Value: slog.StringValue(attrValue),
		},
	}
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	record = record.Clone()
	record.AddAttrs(attrs...)

	return s.Next.Handle(ctx, record)
}

func (s ContextCauseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextCauseHandler{
		Next: s.Next.WithAttrs(attrs),
	}
}

func (s ContextCauseHandler) WithGroup(name string) slog.Handler {
	return ContextCauseHandler{
		Next: s.Next.WithGroup(name),
	}
}
