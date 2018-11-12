package welcemail

import (
	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
)

type Sender interface {
	Send(email string, userProfile userprofile.UserProfile) error
}

type DefaultSender struct {
	Config config.WelcomeEmailConfiguration
	Dialer *gomail.Dialer
}

func NewDefaultSender(config config.WelcomeEmailConfiguration, dialer *gomail.Dialer) Sender {
	return &DefaultSender{
		Config: config,
		Dialer: dialer,
	}
}

func (d *DefaultSender) Send(email string, userProfile userprofile.UserProfile) error {
	sendReq := mail.SendRequest{
		Dialer:      d.Dialer,
		Sender:      d.Config.Sender,
		SenderName:  d.Config.SenderName,
		Recipient:   email,
		Subject:     d.Config.Subject,
		ReplyTo:     d.Config.ReplyTo,
		ReplyToName: d.Config.ReplyToName,
		// TODO: read email text body from template
		TextBody: "TODO",
		// TODO: read email html body from template
		HTMLBody: "<h1>TODO</h1>",
	}

	return sendReq.Execute()
}
