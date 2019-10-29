package welcemail

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
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
	) error
}

type DefaultTestSender struct {
	AppName   string
	URLPrefix *url.URL
	Config    config.WelcomeEmailConfiguration
	Sender    mail.Sender
}

func NewDefaultTestSender(config config.TenantConfiguration, urlPrefix *url.URL, sender mail.Sender) TestSender {
	return &DefaultTestSender{
		AppName:   config.AppName,
		URLPrefix: urlPrefix,
		Config:    config.UserConfig.WelcomeEmail,
		Sender:    sender,
	}
}

func (d *DefaultTestSender) Send(
	email string,
	textTemplate string,
	htmlTemplate string,
	subject string,
	sender string,
	replyTo string,
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
		"user":       userProfile,
		"url_prefix": d.URLPrefix.String(),
	}

	var textBody string
	if textBody, err = template.ParseTextTemplate("test-text", textTemplate, context); err != nil {
		err = errors.Newf("failed to render test text welcome email: %w", err)
		return
	}

	var htmlBody string
	if htmlBody, err = template.ParseHTMLTemplate("test-html", htmlTemplate, context); err != nil {
		err = errors.Newf("failed to render test HTML welcome email: %w", err)
		return
	}

	err = d.Sender.Send(mail.SendOptions{
		Sender:    check(sender, sender, d.Config.Sender),
		Recipient: email,
		Subject:   check(subject, subject, d.Config.Subject),
		ReplyTo:   check(replyTo, replyTo, d.Config.ReplyTo),
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	})
	if err != nil {
		err = errors.Newf("failed to send test welcome email: %w", err)
	}

	return
}
