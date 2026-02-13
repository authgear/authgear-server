package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebConsentHTML = template.RegisterHTML(
	"web/authflowv2/consent.html",
	Components...,
)
