package config

type DatabaseEnvironmentConfig struct {
	DatabaseURL            string `envconfig:"URL"`
	DatabaseSchema         string `envconfig:"SCHEMA" default:"public"`
	MaxOpenConn            int    `envconfig:"MAX_OPEN_CONN" default:"2"`
	MaxIdleConn            int    `envconfig:"MAX_IDLE_CONN" default:"2"`
	ConnMaxLifetimeSeconds int    `envconfig:"CONN_MAX_LIFETIME" default:"1800"`
	ConnMaxIdleTimeSeconds int    `envconfig:"CONN_MAX_IDLE_TIME" default:"300"`
}
