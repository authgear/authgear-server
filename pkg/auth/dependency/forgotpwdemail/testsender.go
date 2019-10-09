package forgotpwdemail

import (
	"fmt"
	"time"

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
	) error
}

type DefaultTestSender struct {
	Config config.ForgotPasswordConfiguration
	Sender mail.Sender
}

func NewDefaultTestSender(config config.TenantConfiguration, sender mail.Sender) TestSender {
	return &DefaultTestSender{
		Config: config.UserConfig.ForgotPassword,
		Sender: sender,
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
	expireAt :=
		time.Now().UTC().
			Truncate(time.Second * 1).
			Add(time.Second * time.Duration(d.Config.ResetURLLifetime))
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
		"appname": d.Config.AppName,
		"link": fmt.Sprintf(
			"%s/example-reset-password-link",
			d.Config.URLPrefix,
		),
		"email":      email,
		"user_id":    userProfile.ID,
		"user":       userProfile,
		"url_prefix": d.Config.URLPrefix,
		"code":       "dummy-reset-code",
		"expire_at":  expireAt,
	}

	var textBody string
	if textBody, err = template.ParseTextTemplate("test-text", textTemplate, context); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = template.ParseHTMLTemplate("test-html", htmlTemplate, context); err != nil {
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

	return
}
