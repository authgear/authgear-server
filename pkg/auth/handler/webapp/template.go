package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var Components = web.ComponentsHTML

var plainTextComponents = web.ComponentsPlainText

// NOTE: To resolve import cycle in panic_middleware.go, put it here as workaround
var TemplateV2WebFatalErrorHTML = template.RegisterHTML(
	"web/authflowv2/fatal_error.html",
	Components...,
)

var TemplateCSRFErrorHTML = template.RegisterHTML(
	"web/authflowv2/csrf_error_page.html",
	Components...,
)

var TemplateCSRFErrorInstructionHTML = template.RegisterHTML(
	"web/authflowv2/csrf_error_instruction.html",
	Components...,
)
