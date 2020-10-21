package welcomemessage

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type TranslationService interface {
	AppMetadata() (*translation.AppMetadata, error)
	EmailMessageData(msg *translation.MessageSpec, args interface{}) (*translation.EmailMessageData, error)
	SMSMessageData(msg *translation.MessageSpec, args interface{}) (*translation.SMSMessageData, error)
}

type Provider struct {
	Translation          TranslationService
	WelcomeMessageConfig *config.WelcomeMessageConfig
	TaskQueue            task.Queue
}

func (p *Provider) send(emails []string) error {
	if !p.WelcomeMessageConfig.Enabled {
		return nil
	}

	if p.WelcomeMessageConfig.Destination == config.WelcomeMessageDestinationFirst {
		if len(emails) > 1 {
			emails = emails[0:1]
		}
	}

	if len(emails) <= 0 {
		return nil
	}

	appMeta, err := p.Translation.AppMetadata()
	if err != nil {
		return err
	}

	var emailMessages []mail.SendOptions
	for _, email := range emails {
		data := map[string]interface{}{
			"Email":   email,
			"AppName": appMeta.AppName,
		}

		msg, err := p.Translation.EmailMessageData(messageWelcomeMessage, data)
		if err != nil {
			return err
		}

		emailMessages = append(emailMessages, mail.SendOptions{
			Sender:    msg.Sender,
			ReplyTo:   msg.ReplyTo,
			Subject:   msg.Subject,
			Recipient: email,
			TextBody:  msg.TextBody,
			HTMLBody:  msg.HTMLBody,
		})
	}

	p.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		EmailMessages: emailMessages,
	})

	return nil
}

func (p *Provider) SendToIdentityInfos(infos []*identity.Info) (err error) {
	var emails []string
	for _, info := range infos {
		if email, ok := info.Claims[identity.StandardClaimEmail].(string); ok {
			emails = append(emails, email)
		}
	}
	return p.send(emails)
}
