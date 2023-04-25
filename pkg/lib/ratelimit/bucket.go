package ratelimit

import (
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type BucketSpec struct {
	Name      string
	Arguments []string
	IsGlobal  bool

	Enabled bool
	Period  time.Duration
	Burst   int
}

var BucketSpecDisabled = BucketSpec{Enabled: false}

func NewBucketSpec(config *config.RateLimitConfig, name string, args ...string) BucketSpec {
	enabled := config.Enabled != nil && *config.Enabled
	var duration time.Duration
	if enabled {
		duration = config.Period.Duration()
	}

	return BucketSpec{
		Name:      name,
		Arguments: args,

		Enabled: enabled,
		Period:  duration,
		Burst:   config.Burst,
	}
}

func NewGlobalBucketSpec(e config.RateLimitsEnvironmentConfigEntry, name string, args ...string) BucketSpec {
	return BucketSpec{
		Name:      name,
		Arguments: args,
		IsGlobal:  true,

		Enabled: e.Enabled,
		Period:  e.Period,
		Burst:   e.Burst,
	}
}

func (s BucketSpec) Key() string {
	return strings.Join(append([]string{s.Name}, s.Arguments...), ":")
}
