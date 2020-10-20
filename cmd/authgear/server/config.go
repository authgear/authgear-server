package server

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Config struct {
	// MainListenAddr sets the listen address of the main server
	MainListenAddr string `envconfig:"MAIN_LISTEN_ADDR" default:"0.0.0.0:3000"`
	// ResolverListenAddr sets the listen address of the resolver server
	ResolverListenAddr string `envconfig:"RESOLVER_LISTEN_ADDR" default:"0.0.0.0:3001"`
	// AdminListenAddr sets the listen address of the admin API server
	AdminListenAddr string `envconfig:"ADMIN_LISTEN_ADDR" default:"0.0.0.0:3002"`

	// TLSCertFilePath sets the file path of TLS certificate.
	// It is only used when development mode is enabled.
	TLSCertFilePath string `envconfig:"TLS_CERT_FILE_PATH" default:"tls-cert.pem"`
	// TLSKeyFilePath sets the file path of TLS private key.
	// It is only used when development mode is enabled.
	TLSKeyFilePath string `envconfig:"TLS_KEY_FILE_PATH" default:"tls-key.pem"`

	// AdminAPIAuth indicates the authorization mode of Admin API
	AdminAPIAuth config.AdminAPIAuth `envconfig:"ADMIN_API_AUTH" default:"jwt"`
	// ConfigSource configures the source of app configurations
	ConfigSource *configsource.Config `envconfig:"CONFIG_SOURCE"`

	// BuiltinResourceDirectory sets the directory for built-in resource files
	BuiltinResourceDirectory string `envconfig:"BUILTIN_RESOURCE_DIRECTORY" default:"resources/authgear"`
	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"CUSTOM_RESOURCE_DIRECTORY"`
	// StaticAsset configures serving static asset
	StaticAsset StaticAssetConfig `envconfig:"STATIC_ASSET"`

	*config.EnvironmentConfig
}

func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	ctx := &validation.Context{}

	switch c.AdminAPIAuth {
	case config.AdminAPIAuthNone, config.AdminAPIAuthJWT:
		break
	default:
		ctx.Child("ADMIN_API_AUTH").EmitErrorMessage(
			"invalid admin API auth mode: must be one of 'none' or 'jwt'",
		)
	}

	if c.StaticAsset.ServingEnabled && c.EnvironmentConfig.StaticAssetURLPrefix == "" {
		ctx.Child("STATIC_ASSET_URL_PREFIX").EmitErrorMessage(
			"static asset URL prefix must be set when static assets are not served",
		)
	}

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

	return ctx.Error("invalid server configuration")
}

type StaticAssetConfig struct {
	// ServingEnabled sets whether serving static assets is enabled
	ServingEnabled bool `envconfig:"SERVING_ENABLED" default:"true"`
	// Dir sets the local directory of static assets
	Dir string `envconfig:"DIR" default:"./static"`
}
