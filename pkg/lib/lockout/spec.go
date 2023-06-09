package lockout

import (
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type LockoutSpec struct {
	Name      string
	Arguments []string

	Enabled         bool
	MaxAttempts     int
	HistoryDuration time.Duration
	MinimumDuration time.Duration
	MaximumDuration time.Duration
	BackoffFactor   float64
	IsGlobal        bool
}

func newLockoutSpec(
	name string,
	maxAttempts int,
	historyDuration time.Duration,
	minimumDuration time.Duration,
	maximumDuration time.Duration,
	backoffFactor float64,
	isGlobal bool,
	args ...string) LockoutSpec {
	enabled := maxAttempts > 0

	if !enabled {
		return LockoutSpec{
			Enabled: false,
		}
	}

	return LockoutSpec{
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

func (s LockoutSpec) Key() string {
	return strings.Join(append([]string{string(s.Name)}, s.Arguments...), ":")
}

func NewAccountAuthenticationSpec(cfg *config.AuthenticationLockoutConfig, userID string) LockoutSpec {
	isGlobal := cfg.LockoutType == config.AuthenticationLockoutTypePerUser
	return newLockoutSpec(
		"AccountAuthentication",
		cfg.MaxAttempts,
		cfg.HistoryDuration.Duration(),
		cfg.MinimumDuration.Duration(),
		cfg.MaximumDuration.Duration(),
		cfg.BackoffFactor,
		isGlobal,
		userID,
	)
}
