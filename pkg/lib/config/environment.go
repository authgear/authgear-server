package config

import "strings"

type TrustProxy bool

type DevMode bool

type SentryDSN string

type ImagesCDNHost string

type WebAppCDNHost string

type CORSAllowedOrigins string

type NFTIndexerAPIEndpoint string

type DenoEndpoint string

func (c *CORSAllowedOrigins) List() []string {
	if string(*c) == "" {
		return []string{}
	}

	return strings.Split(string(*c), ",")
}

type EnvironmentConfig struct {
	// TrustProxy sets whether HTTP headers from proxy are to be trusted
	TrustProxy TrustProxy `envconfig:"TRUST_PROXY" default:"false"`
	// DevMode sets whether the server would be run under development mode
	DevMode DevMode `envconfig:"DEV_MODE" default:"false"`
	// LogLevel sets the global log level
	LogLevel string `envconfig:"LOG_LEVEL" default:"warn"`
	// SentryDSN sets the sentry DSN.
	SentryDSN SentryDSN `envconfig:"SENTRY_DSN"`
	// GlobalDatabase configures the global database
	GlobalDatabase GlobalDatabaseCredentialsEnvironmentConfig `envconfig:"DATABASE"`
	// AuditDatabase configures the audit database
	AuditDatabase AuditDatabaseCredentialsEnvironmentConfig `envconfig:"AUDIT_DATABASE"`
	// DatabaseConfig configures the database connection config
	DatabaseConfig DatabaseEnvironmentConfig `envconfig:"DATABASE_CONFIG"`

	GlobalRedis GlobalRedisCredentialsEnvironmentConfig `envconfig:"REDIS"`
	// RedisConfig configures the redis connection config
	RedisConfig RedisEnvironmentConfig `envconfig:"REDIS_CONFIG"`

	ImagesCDNHost ImagesCDNHost `envconfig:"IMAGES_CDN_HOST"`
	WebAppCDNHost WebAppCDNHost `envconfig:"WEB_APP_CDN_HOST"`

	// CORSAllowOrigins configures a comma-separated list of allowed origins for CORSMiddleware
	CORSAllowedOrigins CORSAllowedOrigins `envconfig:"CORS_ALLOWED_ORIGINS"`

	NFTIndexerAPIEndpoint NFTIndexerAPIEndpoint `envconfig:"NFT_INDEXER_API_ENDPOINT"`

	DenoEndpoint DenoEndpoint `envconfig:"DENO_ENDPOINT"`
}
