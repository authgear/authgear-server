package e2e

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Config struct {
	*config.EnvironmentConfig
}

func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	return cfg, nil
}
