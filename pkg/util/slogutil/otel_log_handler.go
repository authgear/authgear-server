package slogutil

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log"
)

func NewOTelLogHandler(lp log.LoggerProvider, strLevel string) slog.Handler {
	level := ParseLevel(strLevel)
	otelHandler := otelslog.NewHandler(
		"",
		otelslog.WithLoggerProvider(lp),
	)
	return &otelLevelHandler{
		level:   level,
		Handler: otelHandler,
	}
}

type otelLevelHandler struct {
	level slog.Level
	slog.Handler
}

func (h *otelLevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *otelLevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.Handler.Handle(ctx, r)
}

func (h *otelLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelLevelHandler{
		level:   h.level,
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h *otelLevelHandler) WithGroup(name string) slog.Handler {
	return &otelLevelHandler{
		level:   h.level,
		Handler: h.Handler.WithGroup(name),
	}
}
