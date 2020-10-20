package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var (
	TemplateMessageForgotPasswordSMSTXT    = template.RegisterPlainText("messages/forgot_password_sms.txt")
	TemplateMessageForgotPasswordEmailTXT  = template.RegisterPlainText("messages/forgot_password_email.txt")
	TemplateMessageForgotPasswordEmailHTML = template.RegisterHTML("messages/forgot_password_email.html")
)

var messageForgotPassword = &translation.MessageSpec{
	Name:              "forgot-password",
	TXTEmailTemplate:  TemplateMessageForgotPasswordEmailTXT,
	HTMLEmailTemplate: TemplateMessageForgotPasswordEmailHTML,
	SMSTemplate:       TemplateMessageForgotPasswordSMSTXT,
}
