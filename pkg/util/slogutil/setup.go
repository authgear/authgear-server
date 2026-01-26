package slogutil

import (
	"context"
	"log/slog"

	"github.com/kelseyhightower/envconfig"

	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

const (
	LogHandlerConsole = "console"
	LogHandlerOTLP    = "otlp"
)

var ALLOWED_LOG_HANDLERS = []string{
	LogHandlerConsole,
	LogHandlerOTLP,
}

func MakeLogger(ctx context.Context, cfg *LogEnvironmentConfig) *slog.Logger {
	handlers := []slog.Handler{}

	for _, h := range cfg.Handlers {
		switch h {
		case LogHandlerConsole:
			level := cfg.ConsoleLevel
			if level == "" {
				level = cfg.Level
			}
			handlers = append(handlers, NewStderrHandler(level))
		case LogHandlerOTLP:
			if lp := otelutil.GetOTelLoggerProvider(ctx); lp != nil {
				level := cfg.OTLPLevel
				if level == "" {
					level = cfg.Level
				}
				handlers = append(handlers, NewOTelLogHandler(lp, level))
			}
		}
	}

	handlers = append(handlers, NewSentryHandler())

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

// MakeLoggerFromEnv reads environment variables and sets up slog logging.
func MakeLoggerFromEnv(ctx context.Context) *slog.Logger {
	cfg := &LogEnvironmentConfig{}
	_ = envconfig.Process("LOG", cfg)
	return MakeLogger(ctx, cfg)
}

func Setup(ctx context.Context) context.Context {
	logger := MakeLoggerFromEnv(ctx)
	return SetContextLogger(ctx, logger)
}
