package slogutil

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log"
)

func NewOTelLogHandler(lp log.LoggerProvider) slog.Handler {
	otelHandler := otelslog.NewHandler(
		"",
		otelslog.WithLoggerProvider(lp),
	)
	return otelHandler
}
