package server

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Config struct {
	// ListenAddr sets the listen address of the portal server.
	PortalListenAddr string `envconfig:"PORTAL_LISTEN_ADDR" default:"0.0.0.0:3003"`
	// ConfigSource configures the source of app configurations
	ConfigSource *configsource.Config `envconfig:"CONFIG_SOURCE"`
	// Authgear configures Authgear acting as authentication server for the portal.
	Authgear portalconfig.AuthgearConfig `envconfig:"AUTHGEAR"`

	*config.EnvironmentConfig
}

func LoadConfigFromEnv() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	ctx := &validation.Context{}

	switch c.ConfigSource.Type {
	case configsource.TypeLocalFS:
		break
	default:
		sourceTypes := make([]string, len(configsource.Types))
		for i, t := range configsource.Types {
			sourceTypes[i] = string(t)
		}
		ctx.Child("CONFIG_SOURCE_TYPE").EmitErrorMessage(
			"invalid configuration source type; available: " + strings.Join(sourceTypes, ", "),
		)
	}

	if c.Authgear.ClientID == "" {
		ctx.Child("AUTHGEAR_CLIENT_ID").EmitErrorMessage("missing authgear client ID")
	}
	if c.Authgear.Endpoint == "" {
		ctx.Child("AUTHGEAR_ENDPOINT").EmitErrorMessage("missing authgear endpoint")
	}

	return ctx.Error("invalid server configuration")
}
