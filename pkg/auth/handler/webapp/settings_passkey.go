package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsPasskeyHTML = template.RegisterHTML(
	"web/settings_passkey.html",
	Components...,
)

type SettingsPasskeyViewModel struct {
	PasskeyIdentities []*identity.Info
}
