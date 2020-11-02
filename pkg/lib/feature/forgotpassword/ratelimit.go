package forgotpassword

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func GenerateRateLimitBucket(loginID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("reset-password-generate-code:%s", loginID),
		Size:        10,
		ResetPeriod: 1 * time.Minute,
	}
}

func VerifyIPRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("reset-password-verify-ip:%s", ip),
		Size:        10,
		ResetPeriod: 1 * time.Minute,
	}
}
