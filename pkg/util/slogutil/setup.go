package slogutil

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/api/logging"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

func MakeLogger(ctx context.Context, cfg *logging.LogEnvironmentConfig) *slog.Logger {
	handlers := []slog.Handler{}
	otlpRequested := false
	otlpAttached := false

	for _, h := range cfg.Handlers {
		switch h {
		case logging.LogHandlerConsole:
			level := cfg.ConsoleLevel
			if level == "" {
				level = cfg.Level
			}
			handlers = append(handlers, NewStderrHandler(level))
		case logging.LogHandlerOTLP:
			otlpRequested = true
			if lp := otelutil.GetOTelLoggerProvider(ctx); lp != nil {
				level := cfg.OTLPLevel
				if level == "" {
					level = cfg.Level
				}
				handlers = append(handlers, NewOTelLogHandler(lp, level))
				otlpAttached = true
			}
		}
	}

	if otlpRequested && !otlpAttached {
		panic(fmt.Errorf(
			"LOG_HANDLERS includes %q but OTel logger provider is unavailable (handlers=%v level=%q otlp_endpoint=%q); ensure otelutil.SetupOTelSDKGlobally runs before slogutil.Setup and the returned context is preserved",
			logging.LogHandlerOTLP,
			cfg.Handlers.List(),
			cfg.Level,
			cfg.OTLPEndpoint,
		))
	}

	log.Printf(
		"slog setup: handlers=%v level=%q otlp_endpoint=%q otlp_attached=%t",
		cfg.Handlers.List(),
		cfg.Level,
		cfg.OTLPEndpoint,
		otlpAttached,
	)

	handlers = append(handlers, NewSentryHandler())

	// This is the main logging pipeline.
	// It includes the middleware to rich the record,
	// and handle sensitive information.
	mainPipeline := slogmulti.Pipe(
		NewErrorDetailMiddleware(),
		NewStackTraceMiddleware(),
		NewContextCauseMiddleware(),
		NewOTelTraceStateMiddleware(),
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
	cfg := logging.LoadConfig()
	return MakeLogger(ctx, cfg)
}

func Setup(ctx context.Context) context.Context {
	logger := MakeLoggerFromEnv(ctx)
	return SetContextLogger(ctx, logger)
}
