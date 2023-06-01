package tasks

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
)

const SendMessages = "SendMessages"

type SendMessagesParam struct {
	EmailMessages    []mail.SendOptions
	SMSMessages      []sms.SendOptions
	WhatsappMessages []whatsapp.SendTemplateOptions
}

func (p *SendMessagesParam) TaskName() string {
	return SendMessages
}
