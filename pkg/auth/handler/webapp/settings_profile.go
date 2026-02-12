package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileHTML = template.RegisterHTML(
	"web/settings_profile.html",
	Components...,
)

func ConfigureSettingsProfileRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/settings/profile")
}
