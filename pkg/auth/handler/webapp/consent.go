package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebConsentHTML = template.RegisterHTML(
	"web/consent.html",
	Components...,
)
