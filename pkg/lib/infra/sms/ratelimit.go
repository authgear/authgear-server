package sms

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

// TODO(rate-limit): allow configuration of bucket size & reset period

func RateLimitBucket(phone string, messageType string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("sms-message:%s:%s", messageType, phone),
		Size:        1,
		ResetPeriod: duration.PerMinute,
	}
}
