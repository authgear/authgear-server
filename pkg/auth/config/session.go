package config

type SessionConfig struct {
	Lifetime            DurationSeconds `json:"lifetime_seconds,omitempty"`
	IdleTimeoutEnabled  bool            `json:"idle_timeout_enabled,omitempty"`
	IdleTimeout         DurationSeconds `json:"idle_timeout_seconds,omitempty"`
	CookieDomain        *string         `json:"cookie_domain,omitempty"`
	CookieNonPersistent bool            `json:"cookie_non_persistent,omitempty"`
}
