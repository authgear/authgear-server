package mail

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

// TODO(rate-limit): allow configuration of bucket size & reset period

func RateLimitBucket(email string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("email-message:%s", email),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
