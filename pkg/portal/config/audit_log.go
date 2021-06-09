package config

type AuditLogConfig struct {
	Enabled bool `envconfig:"ENABLED" default:"false"`
}
