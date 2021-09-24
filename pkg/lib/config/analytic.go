package config

import (
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type AnalyticConfig struct {
	Enabled bool          `envconfig:"ENABLED" default:"false"`
	Epoch   timeutil.Date `envconfig:"EPOCH"`
}
