package welcemail

import (
	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
)

type TestSender interface {
	Send(
		email string,
		textTemplate string,
		htmlTemplate string,
		subject string,
		sender string,
		replyTo string,
		senderName string,
		replyToName string,
	) error
}

type DefaultTestSender struct {
	Config config.WelcomeEmailConfiguration
	Dialer *gomail.Dialer
}

func NewDefaultTestSender(config config.WelcomeEmailConfiguration, dialer *gomail.Dialer) TestSender {
	return &DefaultTestSender{
		Config: config,
		Dialer: dialer,
	}
}

func (d *DefaultTestSender) Send(
	email string,
	textTemplate string,
	htmlTemplate string,
	subject string,
	sender string,
	replyTo string,
	senderName string,
	replyToName string,
) error {
	check := func(test, a, b string) string {
		if test != "" {
			return a
		}

		return b
	}

	sendReq := mail.SendRequest{
		Dialer:      d.Dialer,
		Sender:      check(sender, sender, d.Config.Sender),
		SenderName:  check(sender, senderName, d.Config.SenderName),
		Recipient:   email,
		Subject:     check(subject, subject, d.Config.Subject),
		ReplyTo:     check(replyTo, replyTo, d.Config.ReplyTo),
		ReplyToName: check(replyTo, replyToName, d.Config.ReplyToName),
		// TODO: read email text body from template
		TextBody: textTemplate,
		// TODO: read email html body from template
		HTMLBody: htmlTemplate,
	}

	return sendReq.Execute()
}
