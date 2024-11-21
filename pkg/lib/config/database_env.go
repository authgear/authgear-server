package config

type DatabaseEnvironmentConfig struct {
	// When you change the default value, you also need to change NewDefaultDatabaseEnvironmentConfig.
	MaxOpenConn int `envconfig:"MAX_OPEN_CONN" default:"3"`
	// When you change the default value, you also need to change NewDefaultDatabaseEnvironmentConfig.
	MaxIdleConn            int             `envconfig:"MAX_IDLE_CONN" default:"3"`
	ConnMaxLifetimeSeconds DurationSeconds `envconfig:"CONN_MAX_LIFETIME" default:"1800"`
	ConnMaxIdleTimeSeconds DurationSeconds `envconfig:"CONN_MAX_IDLE_TIME" default:"300"`
	// USE_PREPARED_STATEMENTS is deprecated. It has no effect anymore.
	//UsePreparedStatements  bool            `envconfig:"USE_PREPARED_STATEMENTS" default:"false"`
}

// NewDefaultDatabaseEnvironmentConfig provides default database config.
// When you changes the default values, you also need to change the values in DatabaseEnvironmentConfig.
func NewDefaultDatabaseEnvironmentConfig() *DatabaseEnvironmentConfig {
	return &DatabaseEnvironmentConfig{
		MaxOpenConn:            3,
		MaxIdleConn:            3,
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
