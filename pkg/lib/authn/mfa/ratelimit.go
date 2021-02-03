package mfa

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func RecoveryCodeAuthRateLimitBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-recovery-code:%s", userID),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}

func DeviceTokenAuthRateLimitBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-device-token:%s", userID),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
