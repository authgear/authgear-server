package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileHTML = template.RegisterHTML(
	"web/settings_profile.html",
	Components...,
)
