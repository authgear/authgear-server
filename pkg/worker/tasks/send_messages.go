package tasks

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func ConfigureSendMessagesTask(registry task.Registry, t task.Task) {
	registry.Register(tasks.SendMessages, t)
}

type MailSender interface {
	Send(opts mail.SendOptions) error
}

type SMSClient interface {
	Send(ctx context.Context, opts sms.SendOptions) error
}

type WhatsappSender interface {
	SendTemplate(ctx context.Context, opts *whatsapp.SendTemplateOptions) error
}

type SendMessagesLogger struct{ *log.Logger }

func NewSendMessagesLogger(lf *log.Factory) SendMessagesLogger {
	return SendMessagesLogger{lf.New("send-messages")}
}

type SendMessagesTask struct {
	EmailSender    MailSender
	SMSClient      SMSClient
	WhatsappSender WhatsappSender
	Logger         SendMessagesLogger
}

func (t *SendMessagesTask) Run(ctx context.Context, param task.Param) (err error) {
	taskParam := param.(*tasks.SendMessagesParam)

	for _, emailMessage := range taskParam.EmailMessages {
		err := t.EmailSender.Send(emailMessage)
		if err != nil {
			t.Logger.WithError(err).WithFields(logrus.Fields{
				"email": mail.MaskAddress(emailMessage.Recipient),
			}).Error("failed to send email")
		}
	}

	for _, smsMessage := range taskParam.SMSMessages {
		err := t.SMSClient.Send(ctx, smsMessage)
		if err != nil {
			t.Logger.WithError(err).WithFields(logrus.Fields{
				"phone": phone.Mask(smsMessage.To),
			}).Error("failed to send SMS")
		}
	}

	for _, whatsappMessage := range taskParam.WhatsappMessages {
		m := whatsappMessage
		err := t.WhatsappSender.SendTemplate(ctx, &m)
		if err != nil {
			t.Logger.WithError(err).WithFields(logrus.Fields{
				"phone": phone.Mask(m.To),
			}).Error("failed to send Whatsapp")
		}
	}

	return
}
