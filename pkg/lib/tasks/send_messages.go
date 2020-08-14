package tasks

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
)

const SendMessages = "SendMessages"

type SendMessagesParam struct {
	EmailMessages []mail.SendOptions
	SMSMessages   []sms.SendOptions
}

func (p *SendMessagesParam) TaskName() string {
	return SendMessages
}
