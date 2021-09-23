package config

type AnalyticConfig struct {
	Enabled bool `envconfig:"ENABLED" default:"false"`
	Epoch   Date `envconfig:"EPOCH"`
}
