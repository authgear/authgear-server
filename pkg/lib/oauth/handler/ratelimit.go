package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func NewBucketSpecOAuthTokenPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthTokenGeneralPerIP, ratelimit.RateLimitGroupOAuthTokenGeneral, &config.RateLimitConfig{
		Enabled: func() *bool { var t = true; return &t }(),
		Period:  "1m",
		Burst:   120,
	}, ratelimit.OAuthTokenPerIP, ip)
}

func NewBucketSpecOAuthTokenPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthTokenGeneralPerUser, ratelimit.RateLimitGroupOAuthTokenGeneral, &config.RateLimitConfig{
		Enabled: func() *bool { var t = true; return &t }(),
		Period:  "1m",
		Burst:   60,
	}, ratelimit.OAuthTokenPerUser, userID)
}

// NewBucketSpecOAuthRegisterPerIP and NewBucketSpecOAuthRegisterPerProject
// bound request volume against POST /oauth2/register, which under open
// registration is unauthenticated and writes a database row per call;
// they are not a substitute for the oauth_client_dcr usage limit. Both
// are project-configurable under oauth.dynamic_client_registration.rate_limits
// (see docs/specs/dcr.md's Rate Limits section) — rateLimits is always
// non-nil by the time a request is handled (its own SetDefaults() fills
// in the built-in fallback rates when the project hasn't configured them).
func NewBucketSpecOAuthRegisterPerIP(rateLimits *config.OAuthDynamicClientRegistrationRateLimitsConfig, ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthRegisterPerIP, ratelimit.RateLimitGroupOAuthRegister, rateLimits.PerIP, ratelimit.OAuthRegisterPerIP, ip)
}

func NewBucketSpecOAuthRegisterPerProject(rateLimits *config.OAuthDynamicClientRegistrationRateLimitsConfig) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthRegisterPerProject, ratelimit.RateLimitGroupOAuthRegister, rateLimits.PerProject, ratelimit.OAuthRegisterPerProject)
}
