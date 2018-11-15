package model

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type GearVersion string

// GearVersion constant specific gear version of app
const (
	LiveVersion      GearVersion = "live"
	PreviousVersion  GearVersion = "previous"
	NightlyVersion   GearVersion = "nightly"
	SuspendedVersion GearVersion = "suspended"
)

// App is skygear application
type App struct {
	ID            string
	Name          string
	Config        config.TenantConfiguration
	Plan          Plan
	AuthVersion   GearVersion
	RecordVersion GearVersion
}

// CanAccessGear determine whether the app can access the given gear
func (a *App) CanAccessGear(gear string) bool {
	return a.Plan.CanAccessGear(gear)
}
