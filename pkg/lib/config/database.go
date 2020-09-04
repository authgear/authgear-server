package config

var _ = Schema.Add("DatabaseConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"max_open_connection": { "type": "integer", "minimum": 0 },
		"max_idle_connection": { "type": "integer", "minimum": 0 },
		"idle_connection_timeout_seconds": { "type": "integer", "minimum": 0 },
		"max_connection_lifetime_seconds": { "type": "integer", "minimum": 0 }
	}
}
`)

type DatabaseConfig struct {
	MaxOpenConnection     *int             `json:"max_open_connection,omitempty"`
	MaxIdleConnection     *int             `json:"max_idle_connection,omitempty"`
	IdleConnectionTimeout *DurationSeconds `json:"idle_connection_timeout_seconds,omitempty"`
	MaxConnectionLifetime *DurationSeconds `json:"max_connection_lifetime_seconds,omitempty"`
}

func (c *DatabaseConfig) SetDefaults() {
	if c.MaxOpenConnection == nil {
		c.MaxOpenConnection = newInt(2)
	}

	if c.MaxIdleConnection == nil {
		c.MaxIdleConnection = newInt(2)
	}

	if c.MaxConnectionLifetime == nil {
		// 30 minutes
		t := DurationSeconds(1800)
		c.MaxConnectionLifetime = &t
	}

	if c.IdleConnectionTimeout == nil {
		// 5 minutes
		t := DurationSeconds(300)
		c.IdleConnectionTimeout = &t
	}
}
