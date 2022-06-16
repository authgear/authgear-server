package config

type RedisEnvironmentConfig struct {
	// Now we use redis pubsub, we need to have much greater number of connections.
	// https://redis.io/topics/clients#maximum-number-of-clients
	MaxOpenConnection     int             `envconfig:"MAX_OPEN_CONN" default:"10000"`
	MaxIdleConnection     int             `envconfig:"MAX_IDLE_CONN" default:"2"`
	MaxConnectionLifetime DurationSeconds `envconfig:"MAX_CONN_LIFETIME" default:"900"`
	IdleConnectionTimeout DurationSeconds `envconfig:"IDLE_CONN_TIMEOUT" default:"300"`
}

// NewDefaultRedisEnvironmentConfig provides default redis config
func NewDefaultRedisEnvironmentConfig() *RedisEnvironmentConfig {
	return &RedisEnvironmentConfig{
		MaxOpenConnection:     10000,
		MaxIdleConnection:     2,
		MaxConnectionLifetime: DurationSeconds(900),
		IdleConnectionTimeout: DurationSeconds(300),
	}
}

type GlobalRedisCredentialsEnvironmentConfig struct {
	RedisURL string `envconfig:"URL"`
}
