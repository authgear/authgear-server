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
	// DefaultSessionLifetime is 52 weeks (364 days).
	DefaultSessionLifetime DurationSeconds = 52 * 7 * 86400
	// DefaultSessionIdleTimeout is 30 days.
	DefaultSessionIdleTimeout DurationSeconds = 30 * 86400
	// DefaultSessionIdleTimeoutEnabled is true.
	DefaultSessionIdleTimeoutEnabled bool = true
	// These default configuration offers a relatively long session lifetime, and disallow prolonged inactivity.
	// For reference, the cookie max age on facebook.com and google.com are 1 year and 2 years respectively.

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
		c.Lifetime = DefaultSessionLifetime
	}
	if c.IdleTimeoutEnabled == nil {
		b := DefaultSessionIdleTimeoutEnabled
		c.IdleTimeoutEnabled = &b
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = DefaultSessionIdleTimeout
	}
}
