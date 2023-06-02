package translation

import "github.com/authgear/authgear-server/pkg/util/template"

type MessageSpec struct {
	Name              string
	TXTEmailTemplate  *template.PlainText
	HTMLEmailTemplate *template.HTML
	SMSTemplate       *template.PlainText
	WhatsappTemplate  *template.PlainText
}

func RegisterMessage(msg *MessageSpec) *MessageSpec {
	return msg
}
