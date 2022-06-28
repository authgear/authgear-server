package cobrasentry

import (
	"context"

	"github.com/getsentry/sentry-go"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func WithHub(ctx context.Context, hub *sentry.Hub) context.Context {
	return context.WithValue(ctx, contextKey, hub)
}

func GetHub(ctx context.Context) *sentry.Hub {
	if hub, ok := ctx.Value(contextKey).(*sentry.Hub); ok {
		return hub
	}
	return nil
}
