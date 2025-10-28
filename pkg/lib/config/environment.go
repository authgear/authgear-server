package config

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/util/crypto"
)

type TrustProxy bool

type DevMode bool

type SentryDSN string

type AuthUISentryDSN string

type AuthUIWindowMessageAllowedOrigins []string

type ImagesCDNHost string

type WebAppCDNHost string

type CORSAllowedOrigins string

type DenoEndpoint string

type AppHostSuffixes []string

type AllowedFrameAncestors []string

type GlobalUIImplementation UIImplementation

type GlobalUISettingsImplementation SettingsUIImplementation

type GlobalSearchImplementation SearchImplementation

type GlobalWhatsappAPIType WhatsappAPIType

type SharedAuthgearEndpoint string

func (s AppHostSuffixes) CheckIsDefaultDomain(host string) bool {
	for _, suffix := range s {
		if before, found := strings.CutSuffix(host, suffix); found {
			if !strings.Contains(before, ".") {
				// We have found a proof.
				return true
			}
		}
	}

	return false
}

func (s AppHostSuffixes) ToWhatsappCloudAPIBizOpaqueCallbackData() string {
	if len(s) == 0 {
		return ""
	}

	// Just join it, and then perform SHA256.
	joined := strings.Join([]string(s), ",")
	return crypto.SHA256String(joined)
}

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
	// AuthUISentryDSN sets the sentry DSN for auth ui.
	AuthUISentryDSN AuthUISentryDSN `envconfig:"AUTH_UI_SENTRY_DSN"`
	// Origins that are allowd to post message to authui
	AuthUIWindowMessageAllowedOrigins AuthUIWindowMessageAllowedOrigins `envconfig:"AUTH_UI_WINDOW_MESSAGE_ALLOWED_ORIGINS"`
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

	AllowedFrameAncestors AllowedFrameAncestors `envconfig:"ALLOWED_FRAME_ANCESTORS"`

	// NFT_INDEXER_API_ENDPOINT is deprecated. It is ignored.
	// Deprecated_NFTIndexerAPIEndpoint NFTIndexerAPIEndpoint `envconfig:"NFT_INDEXER_API_ENDPOINT"`

	DenoEndpoint DenoEndpoint `envconfig:"DENO_ENDPOINT"`

	RateLimits RateLimitsEnvironmentConfig `envconfig:"RATE_LIMITS"`

	SAML SAMLEnvironmentConfig `envconfig:"SAML"`

	// AppHostSuffixes originates from the portal config.
	AppHostSuffixes AppHostSuffixes `envconfig:"APP_HOST_SUFFIXES"`

	// End2EndHTTPProxy sets the HTTP proxy for end-to-end tests
	End2EndHTTPProxy string `envconfig:"E2E_HTTP_PROXY"`
	// End2EndTLSCACertFile sets additional CA certificate for end-to-end tests
	End2EndTLSCACertFile string `envconfig:"E2E_TLS_CA_CERT_FILE"`
	// End2EndBotProtection sets mocked endpoints for bot protection providers verification
	End2EndBotProtection End2EndBotProtectionEnvironmentConfig `envconfig:"E2E_BOT_PROTECTION"`
	// End2EndCSRFProtectionDisabled turns off csrf protection
	End2EndCSRFProtectionDisabled bool `envconfig:"E2E_CSRF_PROTECTION_DISABLED"`

	UIImplementation GlobalUIImplementation `envconfig:"UI_IMPLEMENTATION"`

	UISettingsImplementation GlobalUISettingsImplementation `envconfig:"UI_SETTINGS_IMPLEMENTATION"`

	SearchImplementation GlobalSearchImplementation `envconfig:"SEARCH_IMPLEMENTATION" default:"elasticsearch"`

	// PROJECT_WIZARD_IMPLEMENTATION was deprecated.

	WhatsappAPIType GlobalWhatsappAPIType `envconfig:"WHATSAPP_API_TYPE"`

	UserExportObjectStore *UserExportObjectStoreConfig `envconfig:"USEREXPORT_OBJECT_STORE"`

	SMSGatewayConfig SMSGatewayEnvironmentConfig `envconfig:"SMS_GATEWAY"`

	SharedAuthgearEndpoint SharedAuthgearEndpoint `envconfig:"SHARED_AUTHGEAR_ENDPOINT"`
}
