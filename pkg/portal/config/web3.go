package config

type Web3Config struct {
	Enabled bool `envconfig:"ENABLED" default:"false"`
}
