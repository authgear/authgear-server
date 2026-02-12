package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var SettingsChangePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_old_password": { "type": "string" },
			"x_new_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["x_old_password", "x_new_password", "x_confirm_password"]
	}
`)

func ConfigureSettingsChangePasswordRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/change_password")
}
