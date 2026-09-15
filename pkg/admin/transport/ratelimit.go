package transport

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// NewBucketSpecAdminAPIMutationAllPerProject bounds Admin API mutation volume
// (docs/specs/rate-limit.md § Admin API mutations). Every top-level mutation
// field takes one token.
//
// It takes the resolved feature config scope rather than building a
// config.RateLimitConfig literal, so the built-in rate lives in
// AdminAPIRateLimitsMutationScopeFeatureConfig.SetDefaults and is not
// duplicated here -- the same shape as NewBucketSpecOAuthRegisterPerIP.
func NewBucketSpecAdminAPIMutationAllPerProject(rateLimits *config.AdminAPIRateLimitsMutationScopeFeatureConfig) ratelimit.BucketSpec {
	// No args: BucketSpec.IsGlobal is false, so Limiter keys by app id.
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitAdminAPIMutationAllPerProject,
		ratelimit.RateLimitGroupAdminAPIMutation,
		rateLimits.PerProject,
		ratelimit.AdminAPIMutationAllPerProject,
	)
}
