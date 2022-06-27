package usage

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type HardSMSBucketer struct {
	FeatureConfig *config.FeatureConfig
}

func (b *HardSMSBucketer) Bucket() ratelimit.Bucket {
	c := b.FeatureConfig.RateLimit.SMS
	return ratelimit.Bucket{
		Kind:        ratelimit.KindUsage,
		Key:         "sms-message-hard",
		Size:        *c.Size,
		ResetPeriod: c.ResetPeriod.Duration(),
	}
}
