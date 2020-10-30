package mfa

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func RecoveryCodeAuthRateLimitBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-recovery-code:%s", userID),
		Size:        10,
		ResetPeriod: 1 * time.Minute,
	}
}

func DeviceTokenAuthRateLimitBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-device-token:%s", userID),
		Size:        10,
		ResetPeriod: 1 * time.Minute,
	}
}
