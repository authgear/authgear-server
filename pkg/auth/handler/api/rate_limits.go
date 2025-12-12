package api

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func PresignImageUploadRequestBucketSpec(userID string) ratelimit.BucketSpec {
	enabled := true
	return ratelimit.NewBucketSpec("", "", &config.RateLimitConfig{
		Enabled: &enabled,
		Period:  config.DurationString(duration.PerHour.String()),
		Burst:   10,
	}, ratelimit.PresignImageUploadRequestPerUser, userID)
}
