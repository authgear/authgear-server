package config

type DatabaseEnvironmentConfig struct {
	MaxOpenConn            int             `envconfig:"MAX_OPEN_CONN" default:"2"`
	MaxIdleConn            int             `envconfig:"MAX_IDLE_CONN" default:"2"`
	ConnMaxLifetimeSeconds DurationSeconds `envconfig:"CONN_MAX_LIFETIME" default:"1800"`
	ConnMaxIdleTimeSeconds DurationSeconds `envconfig:"CONN_MAX_IDLE_TIME" default:"300"`
}

// NewDefaultDatabaseEnvironmentConfig provides default database config
func NewDefaultDatabaseEnvironmentConfig() *DatabaseEnvironmentConfig {
	return &DatabaseEnvironmentConfig{
		MaxOpenConn:            2,
		MaxIdleConn:            2,
		ConnMaxLifetimeSeconds: 1800,
		ConnMaxIdleTimeSeconds: 300,
	}
}

type GlobalDatabaseCredentialsEnvironmentConfig struct {
	DatabaseURL    string `envconfig:"URL"`
	DatabaseSchema string `envconfig:"SCHEMA" default:"public"`
}

type AuditDatabaseCredentialsEnvironmentConfig struct {
	DatabaseURL    string `envconfig:"URL"`
	DatabaseSchema string `envconfig:"SCHEMA" default:"public"`
}
