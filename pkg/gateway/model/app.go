package model

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// App is skygear application
type App struct {
	ID     string
	Name   string
	Config config.TenantConfiguration
	Plan   Plan
}

// CanAccessGear determine whether the app can access the given gear
func (a *App) CanAccessGear(gear string) bool {
	return a.Plan.CanAccessGear(gear)
}
