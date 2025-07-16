package slogutil

import (
	"context"
	"log/slog"

	sentryslog "github.com/getsentry/sentry-go/slog"
)

// SentryHandler is a wrapper around sentryslog.SentryHandler.
// It respects IsLoggingSkipped().
type SentryHandler struct {
	Next slog.Handler
}

var _ slog.Handler = (*SentryHandler)(nil)

func (h *SentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Next.Enabled(ctx, level)
}

func (h *SentryHandler) Handle(ctx context.Context, record slog.Record) error {
	if IsLoggingSkipped(record) {
		return nil
	}
	return h.Next.Handle(ctx, record)
}

func (h *SentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SentryHandler{
		Next: h.Next.WithAttrs(attrs),
	}
}

func (h *SentryHandler) WithGroup(name string) slog.Handler {
	return &SentryHandler{
		Next: h.Next.WithGroup(name),
	}
}

func NewSentryHandler() *SentryHandler {
	// The context here is not important.
	// If you read the source, the ctx is used to construct sentry.Logger.
	// We do not use sentry.Logger.
	noctx := context.Background()
	options := sentryslog.Option{
		EventLevel: []slog.Level{
			slog.LevelWarn,
			slog.LevelError,
		},
		// Pass an empty slice to disable sentry.Logger.
		LogLevel:  []slog.Level{},
		AddSource: true,
		// It is intentionally that we do not set Hub here.
		// The Sentry SDK is smart enough to use the hub from context.
		// See https://github.com/getsentry/sentry-go/blob/slog/v0.34.1/slog/sentryslog.go#L178
		// Hub: nil,
	}
	sentryslogHandler := options.NewSentryHandler(noctx)
	return &SentryHandler{
		Next: sentryslogHandler,
	}
}
