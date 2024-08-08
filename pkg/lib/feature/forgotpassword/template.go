package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type messageTemplateContext struct {
	AppName  string
	Email    string
	Password string
}

var (
	TemplateMessageSendPasswordToExistingUserTXT       = template.RegisterMessagePlainText("messages/send_password_to_existing_user_email.txt")
	TemplateMessageSendPasswordToExistingUserEmailHTML = template.RegisterMessageHTML("messages/send_password_to_existing_user_email.html")

	TemplateMessageSendPasswordToNewUserTXT       = template.RegisterMessagePlainText("messages/send_password_to_new_user_email.txt")
	TemplateMessageSendPasswordToNewUserEmailHTML = template.RegisterMessageHTML("messages/send_password_to_new_user_email.html")
)

var (
	messageSendPasswordToExistingUser = &translation.MessageSpec{
		Name:              "send-password-to-existing-user",
		TXTEmailTemplate:  TemplateMessageSendPasswordToExistingUserTXT,
		HTMLEmailTemplate: TemplateMessageSendPasswordToExistingUserEmailHTML,
	}
	messageSendPasswordToNewUser = &translation.MessageSpec{
		Name:              "send-password-to-new-user",
		TXTEmailTemplate:  TemplateMessageSendPasswordToNewUserTXT,
		HTMLEmailTemplate: TemplateMessageSendPasswordToNewUserEmailHTML,
	}
)
