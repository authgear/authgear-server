package cimd

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// noopRateLimiter is the default ServiceRateLimiter binding. A later part
// (the CIMD fetch rate limits) replaces this binding with the real bucket
// checks; until then, Service is independently shippable with fetches
// unthrottled.
type noopRateLimiter struct{}

func (noopRateLimiter) CheckFetchAllowed(ctx context.Context) error { return nil }

func ProvideNoopServiceRateLimiter() ServiceRateLimiter { return noopRateLimiter{} }

// noopUsageLimiter is the default ServiceUsageLimiter binding. A later part
// (the CIMD client limit) replaces this binding with the real usage.Limiter;
// until then Service.upsert reads countBefore but never calls this.
type noopUsageLimiter struct{}

func (noopUsageLimiter) CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error {
	return nil
}

func (noopUsageLimiter) ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int) {
}

func ProvideNoopServiceUsageLimiter() ServiceUsageLimiter { return noopUsageLimiter{} }
