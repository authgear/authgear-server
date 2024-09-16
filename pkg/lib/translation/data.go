package translation

import "github.com/authgear/authgear-server/pkg/util/template"

type EmailMessageData struct {
	Sender   string
	ReplyTo  string
	Subject  string
	HTMLBody *template.RenderResult
	TextBody *template.RenderResult
}

type SMSMessageData struct {
	Sender                    string
	Body                      *template.RenderResult
	PreparedTemplateVariables *PreparedTemplateVariables
}

type WhatsappMessageData struct {
	Body *template.RenderResult
}
