package web

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebRecoveryCodeTXT = template.RegisterPlainText("web/__recovery_code.txt")

var TemplateWebDownloadRecoveryCodeTXT = template.RegisterPlainText(
	"web/download_recovery_code.txt",
	ComponentsPlainText...,
)

var ComponentsPlainText = []*template.PlainText{
	TemplateWebRecoveryCodeTXT,
}
