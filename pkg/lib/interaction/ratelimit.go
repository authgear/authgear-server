package interaction

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type OOBType string

const (
	OOBTypeSetupPrimary          OOBType = "setup-primary-oob"
	OOBTypeSetupSecondary        OOBType = "setup-secondary-oob"
	OOBTypeAuthenticatePrimary   OOBType = "authenticate-primary-oob"
	OOBTypeAuthenticateSecondary OOBType = "authenticate-secondary-oob"
)

// TODO(rate-limit): allow configuration of bucket size & reset period

func RequestRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("request:%s", ip),
		Size:        200,
		ResetPeriod: duration.PerMinute,
	}
}

func SignupRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("signup:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}

func AccountEnumerationRateLimitBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("account-enumeration:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}

func SendVerificationCodeRateLimitBucket(target string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("verification-send-code:%s", target),
		Size:        1,
		ResetPeriod: duration.PerMinute,
	}
}

func SendOOBCodeRateLimitBucket(oobType OOBType, target string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("oob-send-code:%s:%s", oobType, target),
		Size:        1,
		ResetPeriod: duration.PerMinute,
	}
}
