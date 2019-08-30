package config

import (
	"errors"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// Configuration is gateway startup configuration
type Configuration struct {
	Standalone                        bool
	StandaloneTenantConfigurationFile string              `envconfig:"STANDALONE_TENANT_CONFIG_FILE" default:"standalone-tenant-config.yaml"`
	Host                              string              `envconfig:"HOST" default:"localhost:3001"`
	ConnectionStr                     string              `envconfig:"DATABASE_URL"`
	Auth                              GearURLConfig       `envconfig:"AUTH"`
	Redis                             redis.Configuration `envconfig:"REDIS"`
}

// ReadFromEnv reads from environment variable and update the configuration.
func (c *Configuration) ReadFromEnv() error {
	logger := logging.LoggerEntry("gateway")
	if err := godotenv.Load(); err != nil {
		logger.WithError(err).Info(
			"Error in loading .env file, continue without .env")
	}
	err := envconfig.Process("", c)
	if err != nil {
		return err
	}
	return nil
}

type GearURLConfig struct {
	Live    string `envconfig:"LIVE_URL"`
	Nightly string `envconfig:"NIGHTLY_URL"`
}

// GetGearURL provide router map
func (c *Configuration) GetGearURL(gear model.Gear, version model.GearVersion) (string, error) {
	var g GearURLConfig
	switch gear {
	case model.AuthGear:
		g = c.Auth
	default:
		return "", errors.New("invalid gear")
	}

	switch version {
	case model.LiveVersion:
		return g.Live, nil
	case model.NightlyVersion:
		return g.Nightly, nil
	default:
		return "", errors.New("gear is suspended")
	}
}
