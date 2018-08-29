package model

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type App struct {
	Config config.TenantConfiguration
	Gears  map[string]bool
}

func (a *App) CanAccessGear(gear string) bool {
	allow, ok := a.Gears[gear]
	return ok && allow
}

func GetApp(domain string) *App {
	return &App{
		Config: config.TenantConfiguration{
			APIKey: "api-key",
			MasterKey: "master-key",
		},
		Gears: map[string]bool{
			"auth": true,
			"cms": false,
		},
	}
}
