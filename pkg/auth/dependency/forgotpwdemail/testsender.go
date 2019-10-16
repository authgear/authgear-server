package forgotpwdemail

import (
	"net/url"
	"path"
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
	Config    config.ForgotPasswordConfiguration
	URLPrefix *url.URL
	Sender    mail.Sender
}

func NewDefaultTestSender(config config.TenantConfiguration, urlPrefix *url.URL, sender mail.Sender) TestSender {
	return &DefaultTestSender{
		Config:    config.UserConfig.ForgotPassword,
		URLPrefix: urlPrefix,
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
	link := *d.URLPrefix
	link.Path = path.Join(link.Path, "_auth/example-reset-password-link")
	context := map[string]interface{}{
		"appname":    d.Config.AppName,
		"link":       link.String(),
		"email":      email,
		"user_id":    userProfile.ID,
		"user":       userProfile,
		"url_prefix": d.URLPrefix.String(),
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
