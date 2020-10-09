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
	// AdminAPI configures how portal interacts with Authgear Admin API.
	AdminAPI portalconfig.AdminAPIConfig `envconfig:"ADMIN_API"`
	// App configures the managed apps.
	App portalconfig.AppConfig `envconfig:"APP"`
	// StaticAsset configures serving static asset
	StaticAsset StaticAssetConfig `envconfig:"STATIC_ASSET"`
	// Database configures the backend database
	Database portalconfig.DatabaseConfig `envconfig:"DATABASE"`

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

	sourceTypes := make([]string, len(configsource.Types))
	ok := false
	for i, t := range configsource.Types {
		if t == c.ConfigSource.Type {
			ok = true
			break
		}
		sourceTypes[i] = string(t)
	}
	if !ok {
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

	if c.Database.DatabaseURL == "" {
		ctx.Child("DATABASE_URL").EmitErrorMessage("missing database URL")
	}

	c.App.Kubernetes.DefaultDomainTLSCert.
		Validate(ctx.Child("APP_KUBERNETES_DEFAULT_DOMAIN_TLS_CERT"))
	c.App.Kubernetes.CustomDomainTLSCert.
		Validate(ctx.Child("APP_KUBERNETES_CUSTOM_DOMAIN_TLS_CERT"))

	return ctx.Error("invalid server configuration")
}

type StaticAssetConfig struct {
	// ServingEnabled sets whether serving static assets is enabled
	ServingEnabled bool `envconfig:"SERVING_ENABLED" default:"false"`
	// Dir sets the local directory of static assets
	Dir string `envconfig:"DIR" default:"./static"`
}
