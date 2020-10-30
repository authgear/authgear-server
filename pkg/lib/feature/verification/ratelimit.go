package verification

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func VerifyRateLimitBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("verify:%s", userID),
		Size:        10,
		ResetPeriod: 1 * time.Minute,
	}
}
