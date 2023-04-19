package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type Purpose string

type Kind interface {
	Purpose() Purpose
	ValidPeriod() time.Duration

	RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec
	RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec
	RateLimitTriggerCooldown(target string) ratelimit.BucketSpec
	RateLimitValidatePerIP(ip string) ratelimit.BucketSpec
	RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec
	RevocationMaxFailedAttempts() int
}
