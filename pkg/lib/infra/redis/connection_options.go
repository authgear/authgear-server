package redis

import "github.com/authgear/authgear-server/pkg/lib/config"

type ConnectionOptions struct {
	RedisURL              string
	MaxOpenConnection     *int
	MaxIdleConnection     *int
	IdleConnectionTimeout *config.DurationSeconds
	MaxConnectionLifetime *config.DurationSeconds
}

func (c *ConnectionOptions) ConnKey() string {
	return c.RedisURL
}
