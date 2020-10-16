package welcomemessage

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var (
	TemplateMessageWelcomeMessageEmailTXT  = template.RegisterPlainText("messages/welcome_message_email.txt")
	TemplateMessageWelcomeMessageEmailHTML = template.RegisterHTML("messages/welcome_message_email.html")
)

var messageWelcomeMessage = &translation.MessageSpec{
	Name:              "welcome-message",
	TXTEmailTemplate:  TemplateMessageWelcomeMessageEmailTXT,
	HTMLEmailTemplate: TemplateMessageWelcomeMessageEmailHTML,
}
