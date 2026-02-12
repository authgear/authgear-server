package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAHTML = template.RegisterHTML(
	"web/settings_mfa.html",
	Components...,
)

func ConfigureSettingsMFARoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa")
}

type SettingsMFAService interface {
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error
}
