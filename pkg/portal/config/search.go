package config

type SearchConfig struct {
	Enabled bool `envconfig:"ENABLED" default:"false"`
}
