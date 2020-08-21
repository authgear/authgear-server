package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ServerConfig struct {
	// ListenAddr sets the listen address of the portal server.
	PortalListenAddr string `envconfig:"PORTAL_LISTEN_ADDR" default:"0.0.0.0:3003"`
	// TrustProxy sets whether HTTP headers from proxy are to be trusted
	TrustProxy bool `envconfig:"TRUST_PROXY" default:"false"`
	// DevMode sets whether the server would be run under development mode
	DevMode bool `envconfig:"DEV_MODE" default:"false"`
	// LogLevel sets the global log level
	LogLevel string `envconfig:"LOG_LEVEL" default:"warn"`
	// ConfigSource configures the source of app configurations
	ConfigSource ConfigurationSourceConfig `envconfig:"CONFIG_SOURCE"`

	// Authgear configures Authgear acting as authentication server for the portal.
	Authgear AuthgearConfig `envconfig:"AUTHGEAR"`

	// SentryDSN sets the sentry DSN.
	SentryDSN string `envconfig:"SENTRY_DSN"`
}

func LoadServerConfigFromEnv() (*ServerConfig, error) {
	config := &ServerConfig{}

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

func (c *ServerConfig) Validate() error {
	ctx := &validation.Context{}

	switch c.ConfigSource.Type {
	case SourceTypeLocalFile:
		break
	default:
		sourceTypes := make([]string, len(SourceTypes))
		for i, t := range SourceTypes {
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

type SourceType string

const (
	SourceTypeLocalFile SourceType = "local_file"
)

var SourceTypes = []SourceType{
	SourceTypeLocalFile,
}

type ConfigurationSourceConfig struct {
	// Type sets the type of configuration source
	Type SourceType `envconfig:"TYPE" default:"local_file"`

	// AppConfigPath sets the path to app configuration YAML file for local file source
	AppConfigPath string `envconfig:"APP_CONFIG_PATH" default:"authgear.yaml"`
	// SecretConfigPath sets the path to secret configuration YAML file for local file source
	SecretConfigPath string `envconfig:"SECRET_CONFIG_PATH" default:"authgear.secrets.yaml"`
}

type AuthgearConfig struct {
	ClientID string `envconfig:"CLIENT_ID"`
	Endpoint string `envconfig:"ENDPOINT"`
}
