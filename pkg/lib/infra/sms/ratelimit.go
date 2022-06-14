package sms

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

func AntiSpamBucket(phone string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("sms-message:%s", phone),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
