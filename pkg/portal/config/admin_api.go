package config

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AdminAPIType string

const (
	AdminAPITypeStatic AdminAPIType = "static"
)

type AdminAPIConfig struct {
	Type AdminAPIType        `envconfig:"TYPE" default:"static"`
	Auth config.AdminAPIAuth `envconfig:"AUTH" default:"jwt"`
	// Endpoint is used in http.Request.URL to connect the server.
	Endpoint string `envconfig:"ENDPOINT" default:"http://localhost:3002"`
	// HostTemplate is used in http.Request.Host for tenant resolution.
	HostTemplate string `envconfig:"HOST_TEMPLATE" default:"localhost:3002"`
}
