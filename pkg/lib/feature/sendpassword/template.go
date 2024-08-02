package sendpassword

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type MessageType string

const (
	MessageTypeChangePassword MessageType = "change-password"
	MessageTypeCreateUser     MessageType = "create-user"
)

type messageTemplateContext struct {
	AppName  string
	Email    string
	Password string
}

var (
	TemplateMessageChangePasswordTXT       = template.RegisterMessagePlainText("messages/change_password_email.txt")
	TemplateMessageChangePasswordEmailHTML = template.RegisterMessageHTML("messages/change_password_email.html")

	TemplateMessageCreateUserTXT       = template.RegisterMessagePlainText("messages/create_user_email.txt")
	TemplateMessageCreateUserEmailHTML = template.RegisterMessageHTML("messages/create_user_email.html")
)

var (
	messageChangePassword = &translation.MessageSpec{
		Name:              "change-password",
		TXTEmailTemplate:  TemplateMessageChangePasswordTXT,
		HTMLEmailTemplate: TemplateMessageChangePasswordEmailHTML,
	}
	messageCreateUser = &translation.MessageSpec{
		Name:              "create-user",
		TXTEmailTemplate:  TemplateMessageCreateUserTXT,
		HTMLEmailTemplate: TemplateMessageCreateUserEmailHTML,
	}
)
