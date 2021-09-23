package config

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AnalyticConfig struct {
	Enabled bool        `envconfig:"ENABLED" default:"false"`
	Epoch   config.Date `envconfig:"EPOCH"`
}
