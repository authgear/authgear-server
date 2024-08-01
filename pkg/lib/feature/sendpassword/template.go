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
)

var (
	messageChangePassword = &translation.MessageSpec{
		Name:              "change-password",
		TXTEmailTemplate:  TemplateMessageChangePasswordTXT,
		HTMLEmailTemplate: TemplateMessageChangePasswordEmailHTML,
	}
	// TODO: Add template for create user
	messageCreateUser = &translation.MessageSpec{
		Name:              "create-user",
		TXTEmailTemplate:  TemplateMessageChangePasswordTXT,
		HTMLEmailTemplate: TemplateMessageChangePasswordEmailHTML,
	}
)
