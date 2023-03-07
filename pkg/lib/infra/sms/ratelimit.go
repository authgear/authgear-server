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

func (m *AntiSpamSMSBucketMaker) IsPerPhoneEnabled() bool {
	return m.Config.Ratelimit.PerPhone.Enabled
}

func (m *AntiSpamSMSBucketMaker) MakePerPhoneBucket(phone string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("sms-message:%s", phone),
		Name:        "AntiSpamSMSBucket",
		Size:        m.Config.Ratelimit.PerPhone.Size,
		ResetPeriod: m.Config.Ratelimit.PerPhone.ResetPeriod.Duration(),
	}
}

func (m *AntiSpamSMSBucketMaker) IsPerIPEnabled() bool {
	return m.Config.Ratelimit.PerIP.Enabled
}

func (m *AntiSpamSMSBucketMaker) MakePerIPBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("sms-message-per-ip:%s", ip),
		Name:        "AntiSpamSMSBucket",
		Size:        m.Config.Ratelimit.PerIP.Size,
		ResetPeriod: m.Config.Ratelimit.PerIP.ResetPeriod.Duration(),
	}
}
