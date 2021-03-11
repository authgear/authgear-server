package config

var _ = Schema.Add("SessionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"lifetime_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"idle_timeout_enabled": { "type": "boolean" },
		"idle_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"cookie_non_persistent": { "type": "boolean" }
	}
}
`)

const (
	// DefaultIDPSessionLifetime is 52 weeks (364 days).
	DefaultIDPSessionLifetime DurationSeconds = 52 * 7 * 86400
	// DefaultIDPSessionIdleTimeout is 30 days.
	DefaultIDPSessionIdleTimeout DurationSeconds = 30 * 86400
	// DefaultIDPSessionIdleTimeoutEnabled is true.
	DefaultIDPSessionIdleTimeoutEnabled bool = true
	// These default configuration offers a relatively long session lifetime, and disallow prolonged inactivity.
	// For reference, the cookie max age on facebook.com and google.com are 1 year and 2 years respectively.

	// DefaultRefreshTokenLifetime is DefaultIDPSessionLifetime.
	DefaultRefreshTokenLifetime DurationSeconds = DefaultIDPSessionLifetime
	// DefaultRefreshTokenIdleTimeout is DefaultIDPSessionIdleTimeout.
	DefaultRefreshTokenIdleTimeout DurationSeconds = DefaultIDPSessionIdleTimeout
	// DefaultRefreshTokenIdleTimeoutEnabled is DefaultIDPSessionIdleTimeoutEnabled.
	DefaultRefreshTokenIdleTimeoutEnabled bool = DefaultIDPSessionIdleTimeoutEnabled

	// DefaultAccessTokenLifetime is 30 minutes.
	DefaultAccessTokenLifetime DurationSeconds = 30 * 60
)

type SessionConfig struct {
	Lifetime            DurationSeconds `json:"lifetime_seconds,omitempty"`
	IdleTimeoutEnabled  *bool           `json:"idle_timeout_enabled,omitempty"`
	IdleTimeout         DurationSeconds `json:"idle_timeout_seconds,omitempty"`
	CookieNonPersistent bool            `json:"cookie_non_persistent,omitempty"`
}

func (c *SessionConfig) SetDefaults() {
	if c.Lifetime == 0 {
		c.Lifetime = DefaultIDPSessionLifetime
	}
	if c.IdleTimeoutEnabled == nil {
		b := DefaultIDPSessionIdleTimeoutEnabled
		c.IdleTimeoutEnabled = &b
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = DefaultIDPSessionIdleTimeout
	}
}
