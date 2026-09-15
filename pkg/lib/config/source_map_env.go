package config

type SourceMapEnvironmentConfig struct {
	// Enabled sets whether source map files (*.map) are served.
	// It is false by default so that the application source code is not disclosed.
	Enabled bool `envconfig:"ENABLED" default:"false"`
	// SentryToken sets the token Sentry sends when it fetches a publicly hosted source map file.
	// When it is non-empty, source map files are protected by HTTP basic authentication,
	// with the token being the password.
	// See https://docs.sentry.io/platforms/javascript/sourcemaps/uploading/hosting-publicly/
	SentryToken string `envconfig:"SENTRY_TOKEN"`
}
