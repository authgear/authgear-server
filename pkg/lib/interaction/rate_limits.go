package interaction

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func SignupPerIPRateLimitBucketSpec(c *config.AuthenticationConfig, isAnonymous bool, ip string) ratelimit.BucketSpec {
	if isAnonymous {
		return ratelimit.NewBucketSpec(c.RateLimits.SignupAnonymous.PerIP, "SignupAnonymousPerIP", ip)
	}
	return ratelimit.NewBucketSpec(c.RateLimits.Signup.PerIP, "SignupPerIP", ip)
}

func AccountEnumerationPerIPRateLimitBucketSpec(c *config.AuthenticationConfig, ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(c.RateLimits.AccountEnumeration.PerIP, "AccountEnumerationPerIP", ip)
}
