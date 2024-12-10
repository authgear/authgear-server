package server

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Config struct {
	*imagesconfig.EnvironmentConfig

	ObjectStore *imagesconfig.ObjectStoreConfig `envconfig:"IMAGES_OBJECT_STORE"`

	// ListenAddr sets the listen address of the images server.
	ListenAddr string `envconfig:"IMAGES_LISTEN_ADDR" default:"0.0.0.0:3004"`

	InternalListenAddr string `envconfig:"IMAGES_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13004"`
}

func LoadConfigFromEnv() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	err = config.Initialize()
	if err != nil {
		return nil, err
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Initialize() error {
	ctx := &validation.Context{}
	objectStoreConfig := config.AbstractObjectStoreConfig(*c.ObjectStore)
	objectStoreConfig.Initialize(ctx.Child("IMAGES_OBJECT_STORE"))
	return ctx.Error("failed to initialize server configuration")
}

func (c *Config) Validate() error {
	ctx := &validation.Context{}
	objectStoreConfig := config.AbstractObjectStoreConfig(*c.ObjectStore)
	objectStoreConfig.Validate(ctx.Child("IMAGES_OBJECT_STORE"))
	return ctx.Error("invalid server configuration")
}
