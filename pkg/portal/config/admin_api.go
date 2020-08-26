package config

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AdminAPIType string

const (
	AdminAPITypeLocal AdminAPIType = "local"
)

type AdminAPIConfig struct {
	Type     AdminAPIType        `envconfig:"TYPE" default:"local"`
	Auth     config.AdminAPIAuth `envconfig:"AUTH" default:"jwt"`
	Endpoint string              `envconfig:"ENDPOINT" default:"http://localhost:3002"`
}
