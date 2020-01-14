package model

import (
	"context"
)

type contextKey string

const contextKeyGatewayContext contextKey = "gateway-context"

type Context struct {
	App        App
	RouteMatch RouteMatch
}

func ContextWithGatewayContext(ctx context.Context, gatewayContext Context) context.Context {
	return context.WithValue(ctx, contextKeyGatewayContext, gatewayContext)
}

func GatewayContextFromContext(ctx context.Context) Context {
	if c, ok := ctx.Value(contextKeyGatewayContext).(Context); ok {
		return c
	}
	return Context{}
}
