package translation

import "github.com/authgear/authgear-server/pkg/util/template"

type MessageSpec struct {
	MessageType       MessageType
	Name              SpecName
	TXTEmailTemplate  *template.MessagePlainText
	HTMLEmailTemplate *template.MessageHTML
	SMSTemplate       *template.MessagePlainText
	WhatsappTemplate  *template.MessagePlainText
}
