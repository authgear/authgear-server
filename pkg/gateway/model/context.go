package model

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type contextKey string

const contextKeyGatewayContext contextKey = "gateway-context"

type Context struct {
	App             App
	DeploymentRoute config.DeploymentRoute
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
