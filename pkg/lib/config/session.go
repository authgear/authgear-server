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

type SessionConfig struct {
	Lifetime            DurationSeconds `json:"lifetime_seconds,omitempty"`
	IdleTimeoutEnabled  bool            `json:"idle_timeout_enabled,omitempty"`
	IdleTimeout         DurationSeconds `json:"idle_timeout_seconds,omitempty"`
	CookieNonPersistent bool            `json:"cookie_non_persistent,omitempty"`
}

func (c *SessionConfig) SetDefaults() {
	if c.Lifetime == 0 {
		c.Lifetime = 86400
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 300
	}
}
