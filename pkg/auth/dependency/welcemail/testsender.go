package welcemail

import (
	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
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
	AppName string
	Config  config.WelcomeEmailConfiguration
	Dialer  *gomail.Dialer
}

func NewDefaultTestSender(config config.TenantConfiguration, dialer *gomail.Dialer) TestSender {
	return &DefaultTestSender{
		AppName: config.AppName,
		Config:  config.WelcomeEmail,
		Dialer:  dialer,
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
) (err error) {
	check := func(test, a, b string) string {
		if test != "" {
			return a
		}

		return b
	}

	userProfile := userprofile.UserProfile{
		ID: "dummy-id",
	}
	context := map[string]interface{}{
		"appname":    d.AppName,
		"email":      userProfile.Data["email"],
		"user_id":    userProfile.ID,
		"user":       userProfile.ToMap(),
		"url_prefix": d.Config.URLPrefix,
	}

	var textBody string
	if textBody, err = template.ParseTextTemplate(textTemplate, context); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = template.ParseHTMLTemplate(htmlTemplate, context); err != nil {
		return
	}

	sendReq := mail.SendRequest{
		Dialer:      d.Dialer,
		Sender:      check(sender, sender, d.Config.Sender),
		SenderName:  check(sender, senderName, d.Config.SenderName),
		Recipient:   email,
		Subject:     check(subject, subject, d.Config.Subject),
		ReplyTo:     check(replyTo, replyTo, d.Config.ReplyTo),
		ReplyToName: check(replyTo, replyToName, d.Config.ReplyToName),
		TextBody:    textBody,
		HTMLBody:    htmlBody,
	}

	err = sendReq.Execute()
	return
}
