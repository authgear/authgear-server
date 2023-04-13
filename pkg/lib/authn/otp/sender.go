package otp

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/messaging"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type SendOptions struct {
	OTP         string
	URL         string
	MessageType MessageType
	OTPMode     OTPMode
}

type EndpointsProvider interface {
	BaseURL() *url.URL
	LoginLinkVerificationEndpointURL() *url.URL
}

type TranslationService interface {
	EmailMessageData(msg *translation.MessageSpec, args interface{}) (*translation.EmailMessageData, error)
	SMSMessageData(msg *translation.MessageSpec, args interface{}) (*translation.SMSMessageData, error)
}

type Sender interface {
	PrepareEmail(email string, msgType nonblocking.MessageType) (*messaging.EmailMessage, error)
	PrepareSMS(phoneNumber string, msgType nonblocking.MessageType) (*messaging.SMSMessage, error)
}

type PreparedMessage struct {
	email *messaging.EmailMessage
	sms   *messaging.SMSMessage
	spec  *translation.MessageSpec
	form  Form
}

func (m *PreparedMessage) Close() {
	if m.email != nil {
		m.email.Close()
	}
	if m.sms != nil {
		m.sms.Close()
	}
}

type MessageSender struct {
	Translation TranslationService
	Endpoints   EndpointsProvider
	Sender      Sender
}

func (s *MessageSender) setupTemplateContext(otp string, msg *PreparedMessage) (*MessageTemplateContext, error) {
	email := ""
	if msg.email != nil {
		email = msg.email.Recipient
	}

	phone := ""
	if msg.sms != nil {
		phone = msg.sms.To
	}

	url := ""
	if msg.form == FormLink {
		linkURL := s.Endpoints.LoginLinkVerificationEndpointURL()
		query := linkURL.Query()
		query.Set("code", otp)
		linkURL.RawQuery = query.Encode()

		url = linkURL.String()
	}

	return &MessageTemplateContext{
		Email: email,
		Phone: phone,
		Code:  otp,
		URL:   url,
		Host:  s.Endpoints.BaseURL().Host,
	}, nil
}

func (s *MessageSender) selectMessage(form Form, typ MessageType) (*translation.MessageSpec, nonblocking.MessageType) {
	var spec *translation.MessageSpec
	var msgType nonblocking.MessageType
	switch typ {
	case MessageTypeVerification:
		spec = messageVerification
		msgType = nonblocking.MessageTypeVerification
	case MessageTypeSetupPrimaryOOB:
		if form == FormLink {
			spec = messageSetupPrimaryLoginLink
		} else {
			spec = messageSetupPrimaryOOB
		}
		msgType = nonblocking.MessageTypeSetupPrimaryOOB
	case MessageTypeSetupSecondaryOOB:
		if form == FormLink {
			spec = messageSetupSecondaryLoginLink
		} else {
			spec = messageSetupSecondaryOOB
		}
		msgType = nonblocking.MessageTypeSetupSecondaryOOB
	case MessageTypeAuthenticatePrimaryOOB:
		if form == FormLink {
			spec = messageAuthenticatePrimaryLoginLink
		} else {
			spec = messageAuthenticatePrimaryOOB
		}
		msgType = nonblocking.MessageTypeAuthenticatePrimaryOOB
	case MessageTypeAuthenticateSecondaryOOB:
		if form == FormLink {
			spec = messageAuthenticateSecondaryLoginLink
		} else {
			spec = messageAuthenticateSecondaryOOB
		}
		msgType = nonblocking.MessageTypeAuthenticateSecondaryOOB
	default:
		panic("otp: unknown message type: " + msgType)
	}

	return spec, msgType
}

func (s *MessageSender) Prepare(channel model.AuthenticatorOOBChannel, target string, form Form, typ MessageType) (*PreparedMessage, error) {
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		return s.prepareEmail(target, form, typ)
	case model.AuthenticatorOOBChannelSMS:
		return s.prepareSMS(target, form, typ)
	default:
		panic("otp: unknown channel: " + channel)
	}
}

func (s *MessageSender) prepareEmail(email string, form Form, typ MessageType) (*PreparedMessage, error) {
	spec, msgType := s.selectMessage(form, typ)

	msg, err := s.Sender.PrepareEmail(email, msgType)
	if err != nil {
		return nil, err
	}

	return &PreparedMessage{
		email: msg,
		spec:  spec,
		form:  form,
	}, nil
}

func (s *MessageSender) prepareSMS(phoneNumber string, form Form, typ MessageType) (*PreparedMessage, error) {
	spec, msgType := s.selectMessage(form, typ)

	msg, err := s.Sender.PrepareSMS(phoneNumber, msgType)
	if err != nil {
		return nil, err
	}

	return &PreparedMessage{
		sms:  msg,
		spec: spec,
		form: form,
	}, nil
}

func (s *MessageSender) Send(msg *PreparedMessage, otp string) error {
	if msg.email != nil {
		return s.sendEmail(msg, otp)
	}
	if msg.sms != nil {
		return s.sendSMS(msg, otp)
	}
	return nil
}

func (s *MessageSender) sendEmail(msg *PreparedMessage, otp string) error {
	ctx, err := s.setupTemplateContext(otp, msg)
	if err != nil {
		return err
	}

	data, err := s.Translation.EmailMessageData(msg.spec, ctx)
	if err != nil {
		return err
	}

	msg.email.Sender = data.Sender
	msg.email.ReplyTo = data.ReplyTo
	msg.email.Subject = data.Subject
	msg.email.TextBody = data.TextBody
	msg.email.HTMLBody = data.HTMLBody

	return msg.email.Send()
}

func (s *MessageSender) sendSMS(msg *PreparedMessage, otp string) error {
	ctx, err := s.setupTemplateContext(otp, msg)
	if err != nil {
		return err
	}

	data, err := s.Translation.SMSMessageData(msg.spec, ctx)
	if err != nil {
		return err
	}

	msg.sms.Sender = data.Sender
	msg.sms.Body = data.Body

	return msg.sms.Send()
}
