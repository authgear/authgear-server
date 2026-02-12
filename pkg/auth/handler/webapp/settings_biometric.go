package webapp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsBiometricHTML = template.RegisterHTML(
	"web/settings_biometric.html",
	Components...,
)

func ConfigureSettingsBiometricRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/biometric")
}

type BiometricIdentity struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

type SettingsBiometricViewModel struct {
	BiometricIdentities []*BiometricIdentity
}
