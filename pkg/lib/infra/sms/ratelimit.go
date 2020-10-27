package sms

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

// TODO(rate-limit): allow configuration of bucket size & reset period

func RateLimitBucket(phone string, messageType string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("sms-message:%s:%s", messageType, phone),
		Size:        1,
		ResetPeriod: 1 * time.Minute,
	}
}
