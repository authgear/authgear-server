package e2e

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

type Config struct {
	// ConfigSource configures the source of app configurations
	ConfigSource *configsource.Config `envconfig:"CONFIG_SOURCE"`

	// BUILTIN_RESOURCE_DIRECTORY is deprecated. It has no effect anymore.

	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"CUSTOM_RESOURCE_DIRECTORY"`

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
