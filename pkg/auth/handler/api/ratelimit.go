package api

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func AntiSpamPresignImagesUploadBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("presign-images-upload:%s", userID),
		Size:        10,
		ResetPeriod: duration.PerHour,
	}
}
