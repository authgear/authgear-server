package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type RateLimitsEnvironmentConfigEntry struct {
	Enabled bool
	Period  time.Duration
	Burst   int
}

var _ envconfig.Setter = &RateLimitsEnvironmentConfigEntry{}

func (e *RateLimitsEnvironmentConfigEntry) Set(value string) error {
	if value == "" {
		e.Enabled = false
		return nil
	}

	burstStr, periodStr, found := strings.Cut(value, "/")
	if !found {
		return fmt.Errorf("invalid rate limit: %s", value)
	}

	burst, err := strconv.Atoi(burstStr)
	if err != nil {
		return fmt.Errorf("invalid burst value: %w", err)
	} else if burst <= 0 {
		return fmt.Errorf("invalid burst value: %d", burst)
	}

	period, err := time.ParseDuration(periodStr)
	if err != nil {
		return fmt.Errorf("invalid period value: %w", err)
	} else if period <= 0 {
		return fmt.Errorf("invalid period value: %s", period)
	}

	e.Enabled = true
	e.Period = period
	e.Burst = burst
	return nil
}

type RateLimitsEnvironmentConfig struct {
	SMS             RateLimitsEnvironmentConfigEntry `envconfig:"SMS"`
	SMSPerIP        RateLimitsEnvironmentConfigEntry `envconfig:"SMS_PER_IP"`
	SMSPerTarget    RateLimitsEnvironmentConfigEntry `envconfig:"SMS_PER_TARGET" default:"50/24h"`
	Email           RateLimitsEnvironmentConfigEntry `envconfig:"EMAIL"`
	EmailPerIP      RateLimitsEnvironmentConfigEntry `envconfig:"EMAIL_PER_IP"`
	EmailPerTarget  RateLimitsEnvironmentConfigEntry `envconfig:"EMAIL_PER_TARGET" default:"50/24h"`
	TaskUserImport  RateLimitsEnvironmentConfigEntry `envconfig:"TASK_USER_IMPORT"`
	TaskUserExport  RateLimitsEnvironmentConfigEntry `envconfig:"TASK_USER_EXPORT"`
	TaskUserReindex RateLimitsEnvironmentConfigEntry `envconfig:"TASK_USER_REINDEX"`
}
