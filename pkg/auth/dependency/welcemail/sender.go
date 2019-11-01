package welcemail

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type Sender interface {
	Send(urlPrefix *url.URL, email string, user model.User) error
}

type DefaultSender struct {
	AppName        string
	Config         config.WelcomeEmailConfiguration
	Sender         mail.Sender
	TemplateEngine *template.Engine
}

func NewDefaultSender(
	config config.TenantConfiguration,
	sender mail.Sender,
	templateEngine *template.Engine,
) Sender {
	return &DefaultSender{
		AppName:        config.AppName,
		Config:         config.UserConfig.WelcomeEmail,
		Sender:         sender,
		TemplateEngine: templateEngine,
	}
}

func (d *DefaultSender) Send(urlPrefix *url.URL, email string, user model.User) (err error) {
	context := map[string]interface{}{
		"appname":    d.AppName,
		"email":      email,
		"user":       user,
		"url_prefix": urlPrefix.String(),
	}

	var textBody string
	if textBody, err = d.TemplateEngine.ParseTextTemplate(
		authTemplate.TemplateNameWelcomeEmailText,
		context,
		template.ParseOption{Required: true},
	); err != nil {
		err = errors.Newf("failed to render text welcome email: %w", err)
		return
	}

	var htmlBody string
	if htmlBody, err = d.TemplateEngine.ParseHTMLTemplate(
		authTemplate.TemplateNameWelcomeEmailHTML,
		context,
		template.ParseOption{Required: false},
	); err != nil {
		err = errors.Newf("failed to render HTML welcome email: %w", err)
		return
	}

	err = d.Sender.Send(mail.SendOptions{
		Sender:    d.Config.Sender,
		Recipient: email,
		Subject:   d.Config.Subject,
		ReplyTo:   d.Config.ReplyTo,
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	})
	if err != nil {
		err = errors.Newf("failed to send welcome email: %w", err)
	}

	return
}
