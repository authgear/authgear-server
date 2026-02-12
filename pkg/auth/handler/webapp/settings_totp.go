package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsTOTPHTML = template.RegisterHTML(
	"web/settings_totp.html",
	Components...,
)

func ConfigureSettingsTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/totp")
}

type SettingsTOTPViewModel struct {
	Authenticators []*authenticator.Info
}
