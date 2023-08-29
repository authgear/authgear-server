package web

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebRecoveryCodeTXT = template.RegisterPlainText("web/__recovery_code.txt")

var ComponentsPlainText = []*template.PlainText{
	TemplateWebRecoveryCodeTXT,
}
