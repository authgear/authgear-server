package config

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type LogLevel string

type EnvironmentConfig struct {
	// TrustProxy sets whether HTTP headers from proxy are to be trusted
	TrustProxy config.TrustProxy `envconfig:"TRUST_PROXY" default:"false"`
	// LogLevel sets the global log level
	LogLevel LogLevel `envconfig:"LOG_LEVEL" default:"warn"`
	// SentryDSN sets the sentry DSN.
	SentryDSN config.SentryDSN `envconfig:"SENTRY_DSN"`
}
