package lockout

import (
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type BucketName string

const (
	BucketNameAccountAuthentication BucketName = "AccountAuthentication"
)

type BucketSpec struct {
	Name      BucketName
	Arguments []string

	Enabled         bool
	MaxAttempts     int
	HistoryDuration time.Duration
	MinimumDuration time.Duration
	MaximumDuration time.Duration
	BackoffFactor   float64
	IsGlobal        bool
}

func NewBucketSpec(
	name BucketName,
	maxAttempts int,
	historyDuration time.Duration,
	minimumDuration time.Duration,
	maximumDuration time.Duration,
	backoffFactor float64,
	isGlobal bool,
	args ...string) BucketSpec {
	enabled := maxAttempts > 0

	if !enabled {
		return BucketSpec{
			Enabled: false,
		}
	}

	return BucketSpec{
		Name:      name,
		Arguments: args,

		Enabled:         true,
		MaxAttempts:     maxAttempts,
		HistoryDuration: historyDuration,
		MinimumDuration: minimumDuration,
		MaximumDuration: maximumDuration,
		BackoffFactor:   backoffFactor,
		IsGlobal:        isGlobal,
	}
}

func (s BucketSpec) Key() string {
	return strings.Join(append([]string{string(s.Name)}, s.Arguments...), ":")
}

func NewAccountAuthenticationBucket(cfg *config.AuthenticationLockoutConfig, userID string) BucketSpec {
	isGlobal := cfg.LockoutType == config.AuthenticationLockoutTypePerUser
	return NewBucketSpec(
		BucketNameAccountAuthentication,
		cfg.MaxAttempts,
		cfg.HistoryDuration.Duration(),
		cfg.MinimumDuration.Duration(),
		cfg.MaximumDuration.Duration(),
		cfg.BackoffFactor,
		isGlobal,
		userID,
	)
}
