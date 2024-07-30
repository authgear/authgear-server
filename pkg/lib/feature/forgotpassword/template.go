package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var (
	TemplateMessageChangePasswordTXT       = template.RegisterMessagePlainText("messages/change_password_email.txt")
	TemplateMessageChangePasswordEmailHTML = template.RegisterMessageHTML("messages/change_password_email.html")
)

var (
	messageChangePassword = &translation.MessageSpec{
		Name:              "change-password",
		TXTEmailTemplate:  TemplateMessageChangePasswordTXT,
		HTMLEmailTemplate: TemplateMessageChangePasswordEmailHTML,
	}
)
