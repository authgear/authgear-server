package slogutil

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/util/panicutil"
)

var noopLogger = slog.New(slog.DiscardHandler)

type contextKeyTypeLogger struct{}

var contextKeyLogger = contextKeyTypeLogger{}

// SetContextLogger sets logger on ctx.
func SetContextLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		panic(fmt.Errorf("logger must not be nil"))
	}
	ctx = context.WithValue(ctx, contextKeyLogger, logger)
	return ctx
}

// GetContextLogger gets logger from context, or returns noopLogger if no logger is found.
func GetContextLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(contextKeyLogger).(*slog.Logger)
	if !ok {
		return noopLogger
	}
	return logger
}

const AttrKeyLogger = "logger"

type LoggerName string

func NewLogger(name string) LoggerName {
	return LoggerName(name)
}

func (n LoggerName) GetLogger(ctx context.Context) NamedLogger {
	name := string(n)
	logger := GetContextLogger(ctx)
	logger = logger.With(slog.String(AttrKeyLogger, name))
	return NamedLogger{
		name:   name,
		logger: logger,
	}
}

type NamedLogger struct {
	name   string
	logger *slog.Logger
}

func (l NamedLogger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.DebugContext(ctx, msg, attrsToAnys(attrs)...)
}

func (l NamedLogger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.ErrorContext(ctx, msg, attrsToAnys(attrs)...)
}

func (l NamedLogger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.InfoContext(ctx, msg, attrsToAnys(attrs)...)
}

func (l NamedLogger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logger.WarnContext(ctx, msg, attrsToAnys(attrs)...)
}

// With is like slog.Logger.With, except that it takes slog.Attr only.
// The rationale is to minimize programming error of malformed key-value pairs.
func (l NamedLogger) With(attrs ...slog.Attr) NamedLogger {
	anys := make([]any, len(attrs))
	for idx, attr := range attrs {
		anys[idx] = attr
	}

	logger := l.logger.With(anys...)
	return NamedLogger{
		name:   l.name,
		logger: logger,
	}
}

// WithError is a shorthand for With(Err(err)).
func (l NamedLogger) WithError(err error) NamedLogger {
	return l.With(Err(err))
}

// WithRecover is a shorthand for WithError(panicutil.MakeError(r)).
func (l NamedLogger) WithRecover(r any) NamedLogger {
	return l.WithError(panicutil.MakeError(r))
}

// WithGroup is intentionally omitted because it is intended for
// passing a *slog.Logger instance to a third party library.
// We do not have that use case at the moment.

func attrsToAnys(attrs []slog.Attr) []any {
	anys := make([]any, len(attrs))
	for idx, attr := range attrs {
		anys[idx] = attr
	}
	return anys
}
