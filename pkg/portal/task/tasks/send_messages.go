package tasks

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/log"
)

const SendMessages = "SendMessages"

type SendMessagesParam struct {
	EmailMessages []mail.SendOptions
}

func (p *SendMessagesParam) TaskName() string {
	return SendMessages
}

func ConfigureSendMessagesTask(registry task.Registry, t task.Task) {
	registry.Register(SendMessages, t)
}

type MailSender interface {
	Send(opts mail.SendOptions) error
}

type SendMessagesLogger struct{ *log.Logger }

func NewSendMessagesLogger(lf *log.Factory) SendMessagesLogger {
	return SendMessagesLogger{lf.New("send-messages")}
}

type SendMessagesTask struct {
	EmailSender MailSender
	Logger      SendMessagesLogger
}

func (t *SendMessagesTask) Run(ctx context.Context, param task.Param) (err error) {
	taskParam := param.(*SendMessagesParam)

	for _, emailMessage := range taskParam.EmailMessages {
		err := t.EmailSender.Send(emailMessage)
		if err != nil {
			t.Logger.WithError(err).WithFields(logrus.Fields{
				"email": mail.MaskAddress(emailMessage.Recipient),
			}).Error("failed to send email")
		}
	}

	return
}
