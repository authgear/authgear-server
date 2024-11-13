package forgotpassword

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type TranslationService interface {
	EmailMessageData(ctx context.Context, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.EmailMessageData, error)
}

type SenderService interface {
	SendEmailInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error
}

type Sender struct {
	AppConfg    *config.AppConfig
	Identities  IdentityService
	Sender      SenderService
	Translation TranslationService
}

func (s *Sender) getEmailList(ctx context.Context, userID string) ([]string, error) {
	infos, err := s.Identities.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var emails []string
	for _, info := range infos {
		if !info.Type.SupportsPassword() {
			continue
		}

		standardClaims := info.IdentityAwareStandardClaims()
		email := standardClaims[model.ClaimEmail]
		if email != "" {
			emails = append(emails, email)
		}
	}

	return emails, nil
}

func (s *Sender) Send(ctx context.Context, userID string, password string, msgType translation.MessageType) error {
	emails, err := s.getEmailList(ctx, userID)
	if err != nil {
		return err
	}

	if len(emails) == 0 {
		return ErrSendPasswordNoTarget
	}

	for _, email := range emails {
		var spec *translation.MessageSpec
		switch msgType {
		case translation.MessageTypeSendPasswordToExistingUser:
			spec = translation.MessageSendPasswordToExistingUser
		case translation.MessageTypeSendPasswordToNewUser:
			spec = translation.MessageSendPasswordToNewUser
		default:
			panic("forgotpassword: unknown message type: " + msgType)
		}

		partialTemplateVariables := &translation.PartialTemplateVariables{
			Email:    email,
			Password: password,
		}

		data, err := s.Translation.EmailMessageData(ctx, spec, partialTemplateVariables)
		if err != nil {
			return err
		}

		mailSendOptions := &mail.SendOptions{
			Sender:    data.Sender,
			ReplyTo:   data.ReplyTo,
			Subject:   data.Subject,
			Recipient: email,
			TextBody:  data.TextBody.String,
			HTMLBody:  data.HTMLBody.String,
		}

		if err := s.Sender.SendEmailInNewGoroutine(ctx, msgType, mailSendOptions); err != nil {
			return err
		}
	}

	return nil
}
