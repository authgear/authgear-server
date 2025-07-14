package smtp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type SendTestEmailOptions struct {
	To           string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPSender   string
}

type MailSender interface {
	PrepareMessage(opts mail.SendOptions) (*gomail.Message, error)
	Send(*gomail.Message) error
}

var ServiceLogger = slogutil.NewLogger("smtp")

type Service struct {
	DevMode    config.DevMode
	MailSender MailSender
}

func (s *Service) SendRealEmail(ctx context.Context, opts mail.SendOptions) (err error) {
	if s.DevMode {
		logger := ServiceLogger.GetLogger(ctx)
		logger.Warn(ctx, "skip sending email in development mode",
			slog.String("recipient", opts.Recipient),
			slog.String("body", opts.TextBody),
			slog.String("sender", opts.Sender),
			slog.String("subject", opts.Subject),
			slog.String("reply_to", opts.ReplyTo))
		return
	}

	message, err := s.MailSender.PrepareMessage(opts)
	if err != nil {
		return
	}

	err = s.MailSender.Send(message)
	if err != nil {
		return
	}

	return
}

func (s *Service) SendTestEmail(ctx context.Context, app *model.App, options SendTestEmailOptions) (err error) {

	dialer := gomail.NewDialer(
		options.SMTPHost,
		options.SMTPPort,
		options.SMTPUsername,
		options.SMTPPassword,
	)
	// Do not set dialer.SSL so that SSL mode is inferred from the given port.

	message := gomail.NewMessage()
	message.SetHeader("From", options.SMTPSender)
	message.SetHeader("To", options.To)
	message.SetHeader("Subject", "[Test] Authgear email")
	message.SetBody("text/plain", fmt.Sprintf("This email was successfully sent from %s", app.ID))

	err = dialer.DialAndSend(message)
	if err != nil {
		return errors.Join(SMTPTestFailed.New(err.Error()), err)
	}

	return
}
