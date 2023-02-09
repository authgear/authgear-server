package service

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type AntiBruteForceAuthenticateBucketMaker struct {
	PasswordConfig *config.AuthenticatorPasswordConfig
}

func (m *AntiBruteForceAuthenticateBucketMaker) MakeBucket(userID string, authType model.AuthenticatorType) ratelimit.Bucket {
	switch authType {
	case model.AuthenticatorTypePassword:
		return ratelimit.Bucket{
			Key:         fmt.Sprintf("auth-secret:%s:%s", string(authType), userID),
			Size:        m.PasswordConfig.Ratelimit.FailedAttempt.Size,
			ResetPeriod: m.PasswordConfig.Ratelimit.FailedAttempt.ResetPeriod.Duration(),
		}
	default:
		return ratelimit.Bucket{
			Key:         fmt.Sprintf("auth-secret:%s:%s", string(authType), userID),
			Size:        10,
			ResetPeriod: duration.PerMinute,
		}
	}
}
