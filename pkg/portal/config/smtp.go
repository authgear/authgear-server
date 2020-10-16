package config

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SMTPConfig struct {
	Host     string          `envconfig:"HOST"`
	Port     int             `envconfig:"PORT"`
	Username string          `envconfig:"USERNAME"`
	Password string          `envconfig:"PASSWORD"`
	Mode     config.SMTPMode `envconfig:"MODE" default:"normal"`
}
