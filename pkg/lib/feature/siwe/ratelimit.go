package siwe

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

func AntiSpamNonceBucket(ip string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("siwe-nonce:%s", ip),
		Size:        10,
		ResetPeriod: duration.PerMinute,
	}
}
