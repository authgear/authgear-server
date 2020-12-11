package config

var _ = Schema.Add("RedisConfig", `
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

type RedisConfig struct {
	MaxOpenConnection     *int             `json:"max_open_connection,omitempty"`
	MaxIdleConnection     *int             `json:"max_idle_connection,omitempty"`
	IdleConnectionTimeout *DurationSeconds `json:"idle_connection_timeout_seconds,omitempty"`
	MaxConnectionLifetime *DurationSeconds `json:"max_connection_lifetime_seconds,omitempty"`
}

func (c *RedisConfig) SetDefaults() {
	// Now we use redis pubsub, we need to have much greater number of connections.
	// https://redis.io/topics/clients#maximum-number-of-clients
	if c.MaxOpenConnection == nil {
		c.MaxOpenConnection = newInt(10000)
	}
	if c.MaxIdleConnection == nil {
		c.MaxIdleConnection = newInt(2)
	}
	if c.IdleConnectionTimeout == nil {
		// 300 seconds
		t := DurationSeconds(300)
		c.IdleConnectionTimeout = &t
	}
	if c.MaxConnectionLifetime == nil {
		// 15 minutes
		t := DurationSeconds(900)
		c.MaxConnectionLifetime = &t
	}
}
