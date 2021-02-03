package service

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func AuthenticateSecretRateLimitBucket(userID string, authType authn.AuthenticatorType) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("auth-secret:%s:%s", string(authType), userID),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
