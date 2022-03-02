package server

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
)

type Config struct {
	*imagesconfig.EnvironmentConfig

	// ListenAddr sets the listen address of the images server.
	ListenAddr string `envconfig:"PORTAL_LISTEN_ADDR" default:"0.0.0.0:3004"`
}

func LoadConfigFromEnv() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	return config, nil
}
