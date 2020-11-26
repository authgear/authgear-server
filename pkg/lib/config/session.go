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

// DefaultSessionLifetime is 7 days.
// This duration is the same as the maximum lifetime of script-writable cookie imposed by Safari.
// https://webkit.org/blog/10218/full-third-party-cookie-blocking-and-more/
const DefaultSessionLifetime DurationSeconds = 7 * 86400

// DefaultAccessTokenLifetime is 30 minutes.
const DefaultAccessTokenLifetime DurationSeconds = 1800

// DefaultSessionIdleTimeout is 5 minutes.
const DefaultSessionIdleTimeout DurationSeconds = 300

type SessionConfig struct {
	Lifetime            DurationSeconds `json:"lifetime_seconds,omitempty"`
	IdleTimeoutEnabled  bool            `json:"idle_timeout_enabled,omitempty"`
	IdleTimeout         DurationSeconds `json:"idle_timeout_seconds,omitempty"`
	CookieNonPersistent bool            `json:"cookie_non_persistent,omitempty"`
}

func (c *SessionConfig) SetDefaults() {
	if c.Lifetime == 0 {
		c.Lifetime = DefaultSessionLifetime
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = DefaultSessionIdleTimeout
	}
}
