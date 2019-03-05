package model

import (
	"context"
	"time"
)

type CloudCode struct {
	ID         string
	CreatedAt  *time.Time
	Version    string
	Path       string
	TargetPath string
}

const contextKeyCloudCode contextKey = "cloudcode"

func ContextWithCloudCode(ctx context.Context, CloudCode *CloudCode) context.Context {
	return context.WithValue(ctx, contextKeyCloudCode, CloudCode)
}

func CloudCodeFromContext(ctx context.Context) *CloudCode {
	return ctx.Value(contextKeyCloudCode).(*CloudCode)
}
