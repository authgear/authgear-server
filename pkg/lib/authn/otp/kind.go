package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type Kind interface {
	Purpose() string
	AllowLookupByCode() bool
	GenerateCode() string
	VerifyCode(input string, expected string) bool
	ValidPeriod() time.Duration

	RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec
	RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec
	RateLimitTriggerCooldown(target string) ratelimit.BucketSpec
	RateLimitValidatePerIP(ip string) ratelimit.BucketSpec
	RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec
	RevocationMaxFailedAttempts() int
}
