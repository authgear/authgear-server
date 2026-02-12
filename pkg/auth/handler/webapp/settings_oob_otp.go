package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsOOBOTPHTML = template.RegisterHTML(
	"web/settings_oob_otp.html",
	Components...,
)

func ConfigureSettingsOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/oob_otp_:channel")
}

type SettingsOOBOTPViewModel struct {
	OOBAuthenticatorType model.AuthenticatorType
	Authenticators       []*authenticator.Info
}
