package service

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func AntiBruteForceAuthenticateBucket(userID string, authType model.AuthenticatorType) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-secret:%s:%s", string(authType), userID),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
