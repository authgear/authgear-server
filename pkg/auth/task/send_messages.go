package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/sms"
)

func AttachSendMessagesTask(
	executor *async.Executor,
	authDependency auth.DependencyMap,
) {
	executor.Register(spec.SendMessagesTaskName, MakeTask(authDependency, newSendMessagesTask))
}

type SendMessagesTask struct {
	EmailSender   mail.Sender
	SMSClient     sms.Client
	LoggerFactory logging.Factory
}

func (t *SendMessagesTask) Run(ctx context.Context, param interface{}) (err error) {
	taskParam := param.(spec.SendMessagesTaskParam)
	logger := t.LoggerFactory.NewLogger("sendmessages")

	for _, emailMessage := range taskParam.EmailMessages {
		err := t.EmailSender.Send(emailMessage)
		if err != nil {
			logger.WithError(err).WithFields(logrus.Fields{
				"email": mail.MaskAddress(emailMessage.Recipient),
			}).Error("failed to send email")
		}
	}

	for _, smsMessage := range taskParam.SMSMessages {
		err := t.SMSClient.Send(smsMessage)
		if err != nil {
			logger.WithError(err).WithFields(logrus.Fields{
				"phone": phone.Mask(smsMessage.To),
			}).Error("failed to send SMS")
		}
	}

	return
}
