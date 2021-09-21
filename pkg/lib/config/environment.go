package config

type TrustProxy bool

type DevMode bool

type SentryDSN string

type StaticAssetURLPrefix string

type EnvironmentConfig struct {
	// TrustProxy sets whether HTTP headers from proxy are to be trusted
	TrustProxy TrustProxy `envconfig:"TRUST_PROXY" default:"false"`
	// DevMode sets whether the server would be run under development mode
	DevMode DevMode `envconfig:"DEV_MODE" default:"false"`
	// LogLevel sets the global log level
	LogLevel string `envconfig:"LOG_LEVEL" default:"warn"`
	// StaticAssetURLPrefix sets the URL prefix for static assets
	StaticAssetURLPrefix StaticAssetURLPrefix `envconfig:"STATIC_ASSET_URL_PREFIX" default:"/static"`
	// SentryDSN sets the sentry DSN.
	SentryDSN SentryDSN `envconfig:"SENTRY_DSN"`
	// Database configures the backend database
	Database DatabaseEnvironmentConfig `envconfig:"DATABASE"`
	// AuditDatabase configures the audit database
	AuditDatabase DatabaseCredentialsEnvironmentConfig `envconfig:"AUDIT_DATABASE"`
}
