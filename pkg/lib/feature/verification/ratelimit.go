package verification

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func VerifyRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("verification-verify-code:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
