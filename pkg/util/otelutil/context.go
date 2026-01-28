package otelutil

import (
	"context"

	"go.opentelemetry.io/otel/log"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	LoggerProvider log.LoggerProvider
}

func WithOTelLoggerProvider(ctx context.Context, lp log.LoggerProvider) context.Context {
	actx := &contextValue{
		LoggerProvider: lp,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func GetOTelLoggerProvider(ctx context.Context) log.LoggerProvider {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	if actx == nil || actx.LoggerProvider == nil {
		return nil
	}
	return actx.LoggerProvider
}
