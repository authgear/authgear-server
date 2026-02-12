package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsPasskeyHTML = template.RegisterHTML(
	"web/settings_passkey.html",
	Components...,
)

func ConfigureSettingsPasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/passkey")
}

type SettingsPasskeyViewModel struct {
	PasskeyIdentities []*identity.Info
}
