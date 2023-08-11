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
	enabled bool,
	maxAttempts int,
	historyDuration time.Duration,
	minimumDuration time.Duration,
	maximumDuration time.Duration,
	backoffFactor float64,
	isGlobal bool,
	args ...string) LockoutSpec {

	return LockoutSpec{
		Name:      name,
		Arguments: args,

		Enabled:         enabled,
		MaxAttempts:     maxAttempts,
		HistoryDuration: historyDuration,
		MinimumDuration: minimumDuration,
		MaximumDuration: maximumDuration,
		BackoffFactor:   backoffFactor,
		IsGlobal:        isGlobal,
	}
}

func newDisabledLockoutSpec() LockoutSpec {
	return LockoutSpec{Enabled: false}
}

func (s LockoutSpec) Key() string {
	return strings.Join(append([]string{string(s.Name)}, s.Arguments...), ":")
}

func NewAccountAuthenticationSpecForCheck(cfg *config.AuthenticationLockoutConfig, userID string) LockoutSpec {
	isGlobal := cfg.LockoutType == config.AuthenticationLockoutTypePerUser
	if !cfg.IsEnabled() {
		return newDisabledLockoutSpec()
	}
	return newLockoutSpec(
		"AccountAuthentication",
		cfg.IsEnabled(),
		cfg.MaxAttempts,
		cfg.HistoryDuration.Duration(),
		cfg.MinimumDuration.Duration(),
		cfg.MaximumDuration.Duration(),
		*cfg.BackoffFactor,
		isGlobal,
		userID,
	)
}

func NewAccountAuthenticationSpecForAttempt(cfg *config.AuthenticationLockoutConfig, userID string, methods []config.AuthenticationLockoutMethod) LockoutSpec {
	enabled := false
	for _, m := range methods {
		switch m {
		case config.AuthenticationLockoutMethodPassword:
			if cfg.Password.Enabled {
				enabled = true
			}
		case config.AuthenticationLockoutMethodOOBOTP:
			if cfg.OOBOTP.Enabled {
				enabled = true
			}
		case config.AuthenticationLockoutMethodTOTP:
			if cfg.Totp.Enabled {
				enabled = true
			}
		case config.AuthenticationLockoutMethodRecoveryCode:
			if cfg.RecoveryCode.Enabled {
				enabled = true
			}
		}
	}
	if !enabled {
		return newDisabledLockoutSpec()
	}

	return NewAccountAuthenticationSpecForCheck(cfg, userID)
}
