package smtp

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

type SendTestEmailOptions struct {
	To           string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

type Service struct {
	Context context.Context
}

func (s *Service) SendTestEmail(app *model.App, options SendTestEmailOptions) (err error) {
	translationService := NewTranslationService(s.Context, app)
	sender, err := translationService.GetSenderForTestEmail()
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
