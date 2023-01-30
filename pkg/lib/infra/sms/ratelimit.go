package sms

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type AntiSpamSMSBucketMaker struct {
	Config *config.SMSConfig
}

func (m *AntiSpamSMSBucketMaker) MakeBucket(phone string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("sms-message:%s", phone),
		Size:        m.Config.Ratelimit.PerPhone.Size,
		ResetPeriod: m.Config.Ratelimit.PerPhone.ResetPeriod.Duration(),
	}
}
