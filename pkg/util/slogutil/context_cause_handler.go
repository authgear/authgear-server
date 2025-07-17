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
	if ctx != nil {
		// If you read the documentation of https://pkg.go.dev/context#Cause
		// You will see
		//
		// - If ctx is not ended, Cause() return nil
		// - If ctx is ended without a cause, it returns ctx.Err()
		// - If ctx is ended with CancelCauseFunc(err), it returns err.
		//
		// We are only interested the cause of the end of the context, thus calling Cause() is enough.
		if cause := context.Cause(ctx); cause != nil {
			record = record.Clone()
			record.AddAttrs(slog.Attr{
				Key:   "context_cause",
				Value: slog.StringValue(cause.Error()),
			})
		}
	}

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
