package slogutil

import (
	"context"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

type OTelTraceStateHandler struct {
	Next slog.Handler
}

var _ slog.Handler = (*OTelTraceStateHandler)(nil)

func NewOTelTraceStateMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &OTelTraceStateHandler{
			Next: next,
		}
	}
}

func (s *OTelTraceStateHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return s.Next.Enabled(ctx, level)
}

func (s *OTelTraceStateHandler) Handle(ctx context.Context, record slog.Record) error {
	m := otelutil.GetAuthgearBaggage(ctx)
	if len(m) > 0 {
		record = record.Clone()
		for k, v := range m {
			record.AddAttrs(slog.String(k, v))
		}
	}

	return s.Next.Handle(ctx, record)
}

func (s *OTelTraceStateHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OTelTraceStateHandler{
		Next: s.Next.WithAttrs(attrs),
	}
}

func (s *OTelTraceStateHandler) WithGroup(name string) slog.Handler {
	return &OTelTraceStateHandler{
		Next: s.Next.WithGroup(name),
	}
}
