package cobrasentry

import (
	"context"

	"github.com/getsentry/sentry-go"
)

func WithHub(ctx context.Context, hub *sentry.Hub) context.Context {
	return sentry.SetHubOnContext(ctx, hub)
}

func GetHub(ctx context.Context) *sentry.Hub {
	return sentry.GetHubFromContext(ctx)
}
