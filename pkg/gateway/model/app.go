package model

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Gear string

// Gear constant specific gear name
const (
	AuthGear  Gear = "auth"
	AssetGear Gear = "asset"
)

type GearSubdomain string

const (
	AuthGearSubdomain  GearSubdomain = "accounts"
	AssetGearSubdomain GearSubdomain = "assets"
)

// GetGear translate the subdomain to gear if necessary
// otherwise return the original string
func GetGear(subdomain string) Gear {
	switch subdomain {
	case string(AuthGearSubdomain):
		return AuthGear
	case string(AssetGearSubdomain):
		return AssetGear
	}

	return Gear(subdomain)
}

type GearVersion string

// GearVersion constant specific gear version of app
const (
	LiveVersion      GearVersion = "live"
	NightlyVersion   GearVersion = "nightly"
	SuspendedVersion GearVersion = "suspended"
)

// App is skygear application
type App struct {
	ID          string
	Name        string
	Config      config.TenantConfiguration
	Plan        Plan
	AuthVersion GearVersion
}

// CanAccessGear determine whether the app can access the given gear
func (a *App) CanAccessGear(gear Gear) bool {
	return a.Plan.CanAccessGear(gear)
}

func (a *App) GetGearVersion(gear Gear) GearVersion {
	switch gear {
	case AuthGear:
		return a.AuthVersion
	case AssetGear:
		return LiveVersion
	default:
		return SuspendedVersion
	}
}
