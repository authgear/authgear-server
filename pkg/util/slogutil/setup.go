package slogutil

import (
	"context"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
)

func MakeLogger(strLevel string) *slog.Logger {
	pipe := slogmulti.Pipe(
		NewStackTraceMiddleware(),
		NewContextCauseMiddleware(),
		NewSkipLoggingMiddleware(),
		NewMaskMiddleware(NewDefaultMaskHandlerOptions()),
	)
	sink := slogmulti.Fanout(
		NewSentryHandler(),
		NewStderrHandler(strLevel),
		TraceContextCanceledHandler{},
	)
	handler := pipe.Handler(sink)
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
