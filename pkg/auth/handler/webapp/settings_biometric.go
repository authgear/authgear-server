package webapp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsBiometricHTML = template.RegisterHTML(
	"web/settings_biometric.html",
	Components...,
)

type BiometricIdentity struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

type SettingsBiometricViewModel struct {
	BiometricIdentities []*BiometricIdentity
}
