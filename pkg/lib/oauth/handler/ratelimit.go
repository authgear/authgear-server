package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func NewBucketSpecOAuthTokenPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthTokenGeneralPerIP, ratelimit.RateLimitOAuthTokenGeneral, &config.RateLimitConfig{
		Enabled: func() *bool { var t = true; return &t }(),
		Period:  "1m",
		Burst:   120,
	}, ratelimit.OAuthTokenPerIP, ip)
}

func NewBucketSpecOAuthTokenPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthTokenGeneralPerUser, ratelimit.RateLimitOAuthTokenGeneral, &config.RateLimitConfig{
		Enabled: func() *bool { var t = true; return &t }(),
		Period:  "1m",
		Burst:   60,
	}, ratelimit.OAuthTokenPerUser, userID)
}
