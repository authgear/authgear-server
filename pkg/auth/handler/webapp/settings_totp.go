package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsTOTPHTML = template.RegisterHTML(
	"web/settings_totp.html",
	Components...,
)

type SettingsTOTPViewModel struct {
	Authenticators []*authenticator.Info
}
