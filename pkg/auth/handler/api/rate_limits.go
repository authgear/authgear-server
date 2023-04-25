package api

import (
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func PresignImageUploadRequestBucketSpec(userID string) ratelimit.BucketSpec {
	return ratelimit.BucketSpec{
		Name:      "PresignImageUploadRequest",
		Arguments: []string{userID},
		Enabled:   true,
		Period:    duration.PerHour,
		Burst:     10,
	}
}
