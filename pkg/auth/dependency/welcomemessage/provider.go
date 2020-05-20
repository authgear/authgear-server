package welcomemessage

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	taskspec "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type Provider struct {
	AppName                     string
	EmailConfig                 config.EmailMessageConfiguration
	WelcomeMessageConfiguration *config.WelcomeMessageConfiguration
	TemplateEngine              *template.Engine
	TaskQueue                   async.Queue
}

func (p *Provider) send(emails []string) (err error) {
	if !p.WelcomeMessageConfiguration.Enabled {
		return
	}

	if p.WelcomeMessageConfiguration.Destination == config.WelcomeMessageDestinationFirst {
		if len(emails) > 1 {
			emails = emails[0:1]
		}
	}

	if len(emails) <= 0 {
		return
	}

	var emailMessages []mail.SendOptions
	for _, email := range emails {
		data := map[string]interface{}{
			"appname": p.AppName,
			"email":   email,
		}

		var textBody string
		textBody, err = p.TemplateEngine.RenderTemplate(
			TemplateItemTypeWelcomeEmailTXT,
			data,
			template.ResolveOptions{},
		)
		if err != nil {
			return
		}

		var htmlBody string
		htmlBody, err = p.TemplateEngine.RenderTemplate(
			TemplateItemTypeWelcomeEmailHTML,
			data,
			template.ResolveOptions{},
		)
		if err != nil {
			return
		}

		emailMessages = append(emailMessages, mail.SendOptions{
			MessageConfig: p.EmailConfig,
			Recipient:     email,
			TextBody:      textBody,
			HTMLBody:      htmlBody,
		})
	}

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: emailMessages,
		},
	})

	return
}

func (p *Provider) SendToIdentityInfos(infos []*identity.Info) (err error) {
	var emails []string
	for _, info := range infos {
		if email, ok := info.Claims[string(metadata.Email)].(string); ok {
			emails = append(emails, email)
		}
	}
	return p.send(emails)
}
