package verification

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func AutiBruteForceVerifyBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("verification-verify-code:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
