package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsDeleteAccountSuccessHTML = template.RegisterHTML(
	"web/settings_delete_account_success.html",
	Components...,
)

func ConfigureSettingsDeleteAccountSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/delete_account/success")
}

type SettingsDeleteAccountSuccessUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SettingsDeleteAccountSuccessAuthenticationInfoService interface {
	Get(ctx context.Context, entryID string) (entry *authenticationinfo.Entry, err error)
}
