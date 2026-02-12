package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsMFAHTML = template.RegisterHTML(
	"web/settings_mfa.html",
	Components...,
)

type SettingsMFAService interface {
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error
}
