package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsDeleteAccountSuccessHTML = template.RegisterHTML(
	"web/settings_delete_account_success.html",
	Components...,
)

type SettingsDeleteAccountSuccessUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SettingsDeleteAccountSuccessAuthenticationInfoService interface {
	Get(ctx context.Context, entryID string) (entry *authenticationinfo.Entry, err error)
}
