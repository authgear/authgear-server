package smtp

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type SendTestEmailOptions struct {
	To           string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

type MailSender interface {
	Send(opts mail.SendOptions) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("smtp")}
}

type Service struct {
	Logger     Logger
	DevMode    config.DevMode
	MailSender MailSender
}

func (s *Service) SendRealEmail(ctx context.Context, opts mail.SendOptions) (err error) {
	if s.DevMode {
		s.Logger.
			WithField("recipient", opts.Recipient).
			WithField("body", opts.TextBody).
			WithField("sender", opts.Sender).
			WithField("subject", opts.Subject).
			WithField("reply_to", opts.ReplyTo).
			Warn("skip sending email in development mode")
		return
	}

	return s.MailSender.Send(opts)
}

func (s *Service) SendTestEmail(ctx context.Context, app *model.App, options SendTestEmailOptions) (err error) {
	translationService := NewTranslationService(app)
	sender, err := translationService.GetSenderForTestEmail(ctx)
	if err != nil {
		return
	}

	dialer := gomail.NewDialer(
		options.SMTPHost,
		options.SMTPPort,
		options.SMTPUsername,
		options.SMTPPassword,
	)
	// Do not set dialer.SSL so that SSL mode is inferred from the given port.

	message := gomail.NewMessage()
	message.SetHeader("From", sender)
	message.SetHeader("To", options.To)
	message.SetHeader("Subject", "[Test] Authgear email")
	message.SetBody("text/plain", fmt.Sprintf("This email was successfully sent from %s", app.ID))

	err = dialer.DialAndSend(message)
	if err != nil {
		return
	}

	return
}
