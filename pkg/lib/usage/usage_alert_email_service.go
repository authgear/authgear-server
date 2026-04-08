package usage

import (
	"context"
	"errors"
	"log/slog"

	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type TranslationService interface {
	EmailMessageData(ctx context.Context, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.EmailMessageData, error)
}

type MailSender interface {
	PrepareMessage(opts mail.SendOptions) (*gomail.Message, error)
	Send(*gomail.Message) error
}

type UsageAlertEmailService interface {
	Send(ctx context.Context, recipients []string, payload *nonblocking.UsageAlertTriggeredEventPayload) error
}

type UsageAlertEmailServiceImpl struct {
	TranslationService TranslationService
	MailSender         MailSender
	DevMode            config.DevMode
}

func (s *UsageAlertEmailServiceImpl) Send(ctx context.Context, recipients []string, payload *nonblocking.UsageAlertTriggeredEventPayload) error {
	if len(recipients) == 0 {
		return nil
	}

	data, err := s.TranslationService.EmailMessageData(ctx, translation.MessageUsageAlert, &translation.PartialTemplateVariables{
		Usage: &translation.UsageAlertTemplateVariables{
			Name:         payload.Usage.Name,
			Action:       payload.Usage.Action,
			Period:       payload.Usage.Period,
			Quota:        payload.Usage.Quota,
			CurrentValue: payload.Usage.CurrentValue,
		},
	})
	if err != nil {
		return err
	}

	if s.DevMode {
		logger := logger.GetLogger(ctx)
		for _, recipient := range recipients {
			logger.With(
				slog.String("message_type", string(translation.MessageTypeUsageAlert)),
				slog.String("recipient", recipient),
				slog.String("body", data.TextBody.String),
				slog.String("sender", data.Sender),
				slog.String("subject", data.Subject),
				slog.String("reply_to", data.ReplyTo),
			).Info(ctx, "usage alert email is suppressed by development mode")
		}
		return nil
	}

	htmlBody := ""
	if data.HTMLBody != nil {
		htmlBody = data.HTMLBody.String
	}

	var errs []error
	for _, recipient := range recipients {
		message, err := s.MailSender.PrepareMessage(mail.SendOptions{
			Sender:    data.Sender,
			ReplyTo:   data.ReplyTo,
			Subject:   data.Subject,
			Recipient: recipient,
			TextBody:  data.TextBody.String,
			HTMLBody:  htmlBody,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if err := s.MailSender.Send(message); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
