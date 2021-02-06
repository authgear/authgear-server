package interaction

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func RequestRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("request:%s", ip),
		Size:        60,
		ResetPeriod: duration.PerMinute,
	}
}

func SignupRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("signup:%s", ip),
		Size:        1,
		ResetPeriod: duration.PerMinute,
	}
}

func AccountEnumerationRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("account-enumeration:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
