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

func AntiSpamRequestBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("request:%s", ip),
		Size:        200,
		ResetPeriod: duration.PerMinute,
	}
}

func AntiSpamSignupBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("signup:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}

func AntiSpamSignupAnonymousBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("signup-anonymous-user:%s", ip),
		Size:        60,
		ResetPeriod: duration.PerHour,
	}
}

func AntiAccountEnumerationBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("account-enumeration:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}

func AntiSpamSendVerificationCodeBucket(target string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("verification-send-code:%s", target),
		Size:        1,
		ResetPeriod: duration.PerMinute,
	}
}

func AntiSpamSendOOBCodeBucket(oobType OOBType, target string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("oob-send-code:%s:%s", oobType, target),
		Size:        1,
		ResetPeriod: duration.PerMinute,
	}
}
