package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/messaging"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type TranslationService interface {
	EmailMessageData(msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.EmailMessageData, error)
}

type SenderService interface {
	PrepareEmail(email string, msgType translation.MessageType) (*messaging.EmailMessage, error)
}

type Sender struct {
	AppConfg    *config.AppConfig
	Identities  IdentityService
	Sender      SenderService
	Translation TranslationService
}

type PreparedMessage struct {
	email   *messaging.EmailMessage
	spec    *translation.MessageSpec
	msgType translation.MessageType
}

func (m *PreparedMessage) Close() {
	if m.email != nil {
		m.email.Close()
	}
}

func (s *Sender) getEmailList(userID string) ([]string, error) {
	infos, err := s.Identities.ListByUser(userID)
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

func (s *Sender) prepareMessage(email string, msgType translation.MessageType) (*PreparedMessage, error) {
	var spec *translation.MessageSpec

	switch msgType {
	case translation.MessageTypeSendPasswordToExistingUser:
		spec = translation.MessageSendPasswordToExistingUser
	case translation.MessageTypeSendPasswordToNewUser:
		spec = translation.MessageSendPasswordToNewUser
	default:
		panic("forgotpassword: unknown message type: " + msgType)
	}

	msg, err := s.Sender.PrepareEmail(email, msgType)
	if err != nil {
		return nil, err
	}

	return &PreparedMessage{
		email:   msg,
		spec:    spec,
		msgType: msgType,
	}, nil
}

func (s *Sender) Send(userID string, password string, msgType translation.MessageType) error {
	emails, err := s.getEmailList(userID)
	if err != nil {
		return err
	}

	if len(emails) == 0 {
		return ErrSendPasswordNoTarget
	}

	for _, email := range emails {
		msg, err := s.prepareMessage(email, msgType)
		if err != nil {
			return err
		}
		defer msg.Close()

		partialTemplateVariables := &translation.PartialTemplateVariables{
			Email:    email,
			Password: password,
		}

		data, err := s.Translation.EmailMessageData(msg.spec, partialTemplateVariables)
		if err != nil {
			return err
		}

		msg.email.Sender = data.Sender
		msg.email.ReplyTo = data.ReplyTo
		msg.email.Subject = data.Subject
		msg.email.TextBody = data.TextBody.String
		msg.email.HTMLBody = data.HTMLBody.String

		if err := msg.email.Send(); err != nil {
			return err
		}
	}

	return nil
}
