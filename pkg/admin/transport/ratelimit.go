package transport

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// NewBucketSpecAdminAPIMutationAllPerProject and
// NewBucketSpecAdminAPIMutationAllPerIP bound Admin API mutation volume
// (docs/specs/rate-limit.md § Admin API mutations). Every top-level mutation
// field takes one token from each.
//
// They take the resolved feature config scope rather than building a
// config.RateLimitConfig literal, so the built-in rates live in
// AdminAPIRateLimitsMutationScopeFeatureConfig.SetDefaults and are not
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

func NewBucketSpecAdminAPIMutationAllPerIP(rateLimits *config.AdminAPIRateLimitsMutationScopeFeatureConfig, ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitAdminAPIMutationAllPerIP,
		ratelimit.RateLimitGroupAdminAPIMutation,
		rateLimits.PerIP,
		ratelimit.AdminAPIMutationAllPerIP,
		ip,
	)
}
