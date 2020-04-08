package welcemail

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/model"
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
	EmailConfig    config.EmailMessageConfiguration
	Sender         mail.Sender
	TemplateEngine *template.Engine
}

func NewDefaultSender(
	c config.TenantConfiguration,
	sender mail.Sender,
	templateEngine *template.Engine,
) Sender {
	return &DefaultSender{
		AppName: c.AppName,
		EmailConfig: config.NewEmailMessageConfiguration(
			c.AppConfig.Messages.Email,
			c.AppConfig.WelcomeEmail.Message,
		),
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
	if textBody, err = d.TemplateEngine.RenderTemplate(
		TemplateItemTypeWelcomeEmailTXT,
		context,
		template.ResolveOptions{Required: true},
	); err != nil {
		err = errors.Newf("failed to render text welcome email: %w", err)
		return
	}

	var htmlBody string
	if htmlBody, err = d.TemplateEngine.RenderTemplate(
		TemplateItemTypeWelcomeEmailHTML,
		context,
		template.ResolveOptions{Required: false},
	); err != nil {
		err = errors.Newf("failed to render HTML welcome email: %w", err)
		return
	}

	err = d.Sender.Send(mail.SendOptions{
		Sender:    d.EmailConfig.Sender(),
		Recipient: email,
		Subject:   d.EmailConfig.Subject(),
		ReplyTo:   d.EmailConfig.ReplyTo(),
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	})
	if err != nil {
		err = errors.Newf("failed to send welcome email: %w", err)
	}

	return
}
