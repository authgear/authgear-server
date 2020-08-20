package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ServerConfig struct {
	// ListenAddr sets the listen address of the main server
	ListenAddr string `envconfig:"LISTEN_ADDR" default:"0.0.0.0:3000"`
	// ResolverListenAddr sets the listen address of the resolver server
	ResolverListenAddr string `envconfig:"RESOLVER_LISTEN_ADDR" default:"0.0.0.0:3001"`
	// AdminListenAddr sets the listen address of the admin API server
	AdminListenAddr string `envconfig:"ADMIN_LISTEN_ADDR" default:"0.0.0.0:3002"`

	// TrustProxy sets whether HTTP headers from proxy are to be trusted
	TrustProxy bool `envconfig:"TRUST_PROXY" default:"false"`
	// DevMode sets whether the server would be run under development mode
	DevMode bool `envconfig:"DEV_MODE" default:"false"`

	// TLSCertFilePath sets the file path of TLS certificate.
	// It is only used when development mode is enabled.
	TLSCertFilePath string `envconfig:"TLS_CERT_FILE_PATH" default:"tls-cert.pem"`
	// TLSKeyFilePath sets the file path of TLS private key.
	// It is only used when development mode is enabled.
	TLSKeyFilePath string `envconfig:"TLS_KEY_FILE_PATH" default:"tls-key.pem"`

	// AdminAPIAuth indicates the authorization mode of Admin API
	AdminAPIAuth AdminAPIAuth `envconfig:"ADMIN_API_AUTH" default:"jwt"`
	// LogLevel sets the global log level
	LogLevel string `envconfig:"LOG_LEVEL" default:"warn"`
	// ConfigSource configures the source of app configurations
	ConfigSource ConfigurationSourceConfig `envconfig:"CONFIG_SOURCE"`

	// DefaultTemplateDirectory sets the directory for default template files
	DefaultTemplateDirectory string `envconfig:"DEFAULT_TEMPLATE_DIRECTORY" default:"templates"`
	// ReservedNameFilePath sets the file path for reserved name list
	ReservedNameFilePath string `envconfig:"RESERVED_NAME_FILE_PATH" default:"reserved_name.txt"`
	// StaticAsset configures serving static asset
	StaticAsset ServerStaticAssetConfig `envconfig:"STATIC_ASSET"`

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

	switch c.AdminAPIAuth {
	case AdminAPIAuthNone, AdminAPIAuthJWT:
		break
	default:
		ctx.Child("ADMIN_API_AUTH").EmitErrorMessage(
			"invalid admin API auth mode: must be one of 'none' or 'jwt'",
		)
	}

	if c.StaticAsset.ServingEnabled && c.StaticAsset.URLPrefix == "" {
		ctx.Child("STATIC_ASSET_URL_PREFIX").EmitErrorMessage(
			"static asset URL prefix must be set when static assets are not served",
		)
	}

	switch c.ConfigSource.Type {
	case SourceTypeLocalFS:
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

	return ctx.Error("invalid server configuration")
}

type AdminAPIAuth string

const (
	AdminAPIAuthNone AdminAPIAuth = "none"
	AdminAPIAuthJWT  AdminAPIAuth = "jwt"
)

type ServerStaticAssetConfig struct {
	// ServingEnabled sets whether serving static assets is enabled
	ServingEnabled bool `envconfig:"SERVING_ENABLED" default:"true"`
	// Dir sets the local directory of static assets
	Dir string `envconfig:"DIR" default:"./static"`
	// URLPrefix sets the URL prefix for static assets
	URLPrefix string `envconfig:"URL_PREFIX" default:"/static"`
}

type SourceType string

const (
	SourceTypeLocalFS SourceType = "local_fs"
)

var SourceTypes = []SourceType{
	SourceTypeLocalFS,
}

type ConfigurationSourceConfig struct {
	// Type sets the type of configuration source
	Type SourceType `envconfig:"TYPE" default:"local_fs"`

	// Watch indicates whether the configuration source would watch for changes and reload automatically
	Watch bool `envconfig:"WATCH" default:"true"`
	// Directory sets the path to app configuration directory file for local FS sources
	Directory string `envconfig:"DIRECTORY" default:"."`
}
