package slogutil

import (
	"context"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
)

func MakeLogger(strLevel string) *slog.Logger {
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
		NewSentryHandler(),
		NewStderrHandler(strLevel),
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
func MakeLoggerFromEnv() *slog.Logger {
	strLevel := os.Getenv("LOG_LEVEL")
	return MakeLogger(strLevel)
}

func Setup(ctx context.Context) context.Context {
	logger := MakeLoggerFromEnv()
	return SetContextLogger(ctx, logger)
}
