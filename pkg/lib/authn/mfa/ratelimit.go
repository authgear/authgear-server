package mfa

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func AutiBruteForceRecoveryCodeBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-recovery-code:%s", userID),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}

func AntiBruteForceDeviceTokenBucket(userID string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-device-token:%s", userID),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
