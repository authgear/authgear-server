package interaction

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func RequestRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("request:%s", ip),
		Size:        60,
		ResetPeriod: 1 * time.Minute,
	}
}

func SignupRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("signup:%s", ip),
		Size:        1,
		ResetPeriod: 1 * time.Minute,
	}
}

func AuthIPRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-ip:%s", ip),
		Size:        10,
		ResetPeriod: 1 * time.Minute,
	}
}
