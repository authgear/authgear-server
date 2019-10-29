package config

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type configContext struct {
	config *TenantConfiguration
}

func WithTenantConfig(ctx context.Context, config *TenantConfiguration) context.Context {
	configCtx, ok := ctx.Value(contextKey).(*configContext)
	if ok {
		configCtx.config = config
		return ctx
	}

	return context.WithValue(ctx, contextKey, &configContext{config})
}

func GetTenantConfig(ctx context.Context) *TenantConfiguration {
	configCtx, ok := ctx.Value(contextKey).(*configContext)
	if !ok {
		return nil
	}
	return configCtx.config
}
