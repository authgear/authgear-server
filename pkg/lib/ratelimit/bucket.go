package ratelimit

import (
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type BucketName string

type BucketSpec struct {
	Name           BucketName
	RateLimitName  RateLimitName
	RateLimitGroup RateLimitGroup
	Arguments      []string
	IsGlobal       bool

	Enabled bool
	Period  time.Duration
	Burst   int
}

var BucketSpecDisabled = BucketSpec{Enabled: false}

func NewBucketSpec(rl RateLimitName, group RateLimitGroup, config *config.RateLimitConfig, name BucketName, args ...string) BucketSpec {
	enabled := config.IsEnabled()
	var duration time.Duration
	if enabled {
		duration = config.Period.Duration()
	}

	return BucketSpec{
		Name:           name,
		RateLimitName:  rl,
		RateLimitGroup: group,
		Arguments:      args,

		Enabled: enabled,
		Period:  duration,
		Burst:   config.Burst,
	}
}

func NewCooldownSpec(name BucketName, period time.Duration, args ...string) BucketSpec {
	return BucketSpec{
		Name:      name,
		Arguments: args,
		Enabled:   true,
		Period:    period,
		Burst:     1,
	}
}

func NewGlobalBucketSpec(rl RateLimitGroup, e config.RateLimitsEnvironmentConfigEntry, name BucketName, args ...string) BucketSpec {
	return BucketSpec{
		Name:           name,
		RateLimitGroup: rl,
		Arguments:      args,
		IsGlobal:       true,

		Enabled: e.Enabled,
		Period:  e.Period,
		Burst:   e.Burst,
	}
}

func (s BucketSpec) Key() string {
	return strings.Join(append([]string{string(s.Name)}, s.Arguments...), ":")
}
