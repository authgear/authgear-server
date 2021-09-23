package config

type AnalyticConfig struct {
	Enabled bool `envconfig:"ENABLED" default:"false"`
}
