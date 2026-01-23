package slogutil

import (
	"context"
	"log/slog"
	"os"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
	slogmulti "github.com/samber/slog-multi"
)

func MakeLogger(ctx context.Context, strLevel string) *slog.Logger {
	handlers := []slog.Handler{}

	if lp := otelutil.GetOTelLoggerProvider(ctx); lp != nil {
		handlers = append(handlers, NewOTelLogHandler(lp))
	}
	handlers = append(handlers, NewSentryHandler())
	handlers = append(handlers, NewStderrHandler(strLevel))

	// This is the main logging pipeline.
	// It includes the middleware to rich the record,
	// and handle sensitive information.
	mainPipeline := slogmulti.Pipe(
		NewErrorDetailMiddleware(),
		NewStackTraceMiddleware(),
		NewContextCauseMiddleware(),
		NewSkipLoggingMiddleware(),
		NewMaskMiddleware(NewDefaultMaskHandlerOptions()),
	).Handler(slogmulti.Fanout(
		handlers...,
	))

	// The actual handler is a fanout to
	// - a handler that converts error to metric whenever appropriate.
	// - the main handler.
	handler := slogmulti.Fanout(
		NewOtelMetricHandler(),
		mainPipeline,
	)

	logger := slog.New(handler)
	return logger
}

// MakeLoggerFromEnv reads environment variable LOG_LEVEL and sets up slog logging.
// For simplicity, we read the environment variable directly.
func MakeLoggerFromEnv(ctx context.Context) *slog.Logger {
	strLevel := os.Getenv("LOG_LEVEL")
	return MakeLogger(ctx, strLevel)
}

func Setup(ctx context.Context) context.Context {
	logger := MakeLoggerFromEnv(ctx)
	return SetContextLogger(ctx, logger)
}
