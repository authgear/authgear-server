package otp

import (
	"errors"
	neturl "net/url"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/messaging"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type SendOptions struct {
	OTP               string
	AdditionalContext any
}

type EndpointsProvider interface {
	Origin() *neturl.URL
	LoginLinkVerificationEndpointURL() *neturl.URL
	ResetPasswordEndpointURL() *neturl.URL
}

type TranslationService interface {
	EmailMessageData(msg *translation.MessageSpec, args interface{}) (*translation.EmailMessageData, error)
	SMSMessageData(msg *translation.MessageSpec, args interface{}) (*translation.SMSMessageData, error)
	WhatsappMessageData(language string, msg *translation.MessageSpec, args interface{}) (*translation.WhatsappMessageData, error)
}

type Sender interface {
	PrepareEmail(email string, msgType nonblocking.MessageType) (*messaging.EmailMessage, error)
	PrepareSMS(phoneNumber string, msgType nonblocking.MessageType) (*messaging.SMSMessage, error)
	PrepareWhatsapp(phoneNumber string, msgType nonblocking.MessageType) (*messaging.WhatsappMessage, error)
}

type WhatsappService interface {
	ResolveOTPTemplateLanguage() (string, error)
	PrepareOTPTemplate(language string, text string, code string) (*whatsapp.PreparedOTPTemplate, error)
	SendTemplate(opts *whatsapp.SendTemplateOptions) error
}

type PreparedMessage struct {
	email    *messaging.EmailMessage
	sms      *messaging.SMSMessage
	whatsapp *messaging.WhatsappMessage
	spec     *translation.MessageSpec
	form     Form
	msgType  nonblocking.MessageType
}

func (m *PreparedMessage) Close() {
	if m.email != nil {
		m.email.Close()
	}
	if m.sms != nil {
		m.sms.Close()
	}
	if m.whatsapp != nil {
		m.whatsapp.Close()
	}
}

type MessageSender struct {
	Translation     TranslationService
	Endpoints       EndpointsProvider
	Sender          Sender
	WhatsappService WhatsappService
}

func (s *MessageSender) setupTemplateContext(msg *PreparedMessage, opts SendOptions) (any, error) {
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
		var linkURL *neturl.URL
		switch msg.msgType {
		case nonblocking.MessageTypeSetupPrimaryOOB,
			nonblocking.MessageTypeSetupSecondaryOOB,
			nonblocking.MessageTypeAuthenticatePrimaryOOB,
			nonblocking.MessageTypeAuthenticateSecondaryOOB:

			linkURL = s.Endpoints.LoginLinkVerificationEndpointURL()
			query := linkURL.Query()
			query.Set("code", opts.OTP)
			linkURL.RawQuery = query.Encode()

		case nonblocking.MessageTypeForgotPassword:

			linkURL = s.Endpoints.ResetPasswordEndpointURL()
			query := linkURL.Query()
			query.Set("code", opts.OTP)
			linkURL.RawQuery = query.Encode()

		default:
			panic("otp: unexpected message type for link: " + msg.msgType)
		}

		url = linkURL.String()
	}

	ctx := make(map[string]any)
	template.Embed(ctx, messageTemplateContext{
		Email: email,
		Phone: phone,
		Code:  opts.OTP,
		URL:   url,
		Link:  url,
		Host:  s.Endpoints.Origin().Host,
	})
	if opts.AdditionalContext != nil {
		template.Embed(ctx, opts.AdditionalContext)
	}

	return ctx, nil
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
	case MessageTypeForgotPassword:
		if form == FormLink {
			spec = messageForgotPasswordLink
		} else {
			spec = messageForgotPasswordOOB
		}
		msgType = nonblocking.MessageTypeForgotPassword
	case MessageTypeWhatsappCode:
		spec = messageWhatsappCode
		msgType = nonblocking.MessageTypeWhatsappCode
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
	case model.AuthenticatorOOBChannelWhatsapp:
		return s.prepareWhatsapp(target, form, typ)
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
		email:   msg,
		spec:    spec,
		form:    form,
		msgType: msgType,
	}, nil
}

func (s *MessageSender) prepareSMS(phoneNumber string, form Form, typ MessageType) (*PreparedMessage, error) {
	spec, msgType := s.selectMessage(form, typ)

	msg, err := s.Sender.PrepareSMS(phoneNumber, msgType)
	if err != nil {
		return nil, err
	}

	return &PreparedMessage{
		sms:     msg,
		spec:    spec,
		form:    form,
		msgType: msgType,
	}, nil
}

func (s *MessageSender) prepareWhatsapp(phoneNumber string, form Form, typ MessageType) (*PreparedMessage, error) {
	spec, msgType := s.selectMessage(form, typ)

	msg, err := s.Sender.PrepareWhatsapp(phoneNumber, msgType)
	if err != nil {
		return nil, err
	}

	return &PreparedMessage{
		whatsapp: msg,
		spec:     spec,
		form:     form,
		msgType:  msgType,
	}, nil
}

func (s *MessageSender) Send(msg *PreparedMessage, opts SendOptions) error {
	if msg.email != nil {
		return s.sendEmail(msg, opts)
	}
	if msg.sms != nil {
		return s.sendSMS(msg, opts)
	}
	if msg.whatsapp != nil {
		return s.sendWhatsapp(msg, opts)
	}
	return nil
}

func (s *MessageSender) sendEmail(msg *PreparedMessage, opts SendOptions) error {
	ctx, err := s.setupTemplateContext(msg, opts)
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

func (s *MessageSender) sendSMS(msg *PreparedMessage, opts SendOptions) error {
	ctx, err := s.setupTemplateContext(msg, opts)
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

func (s *MessageSender) sendWhatsapp(msg *PreparedMessage, opts SendOptions) (err error) {
	// Rewrite the error to be APIError.
	defer func() {
		if err != nil {
			if errors.Is(err, whatsapp.ErrInvalidUser) {
				err = ErrInvalidWhatsappUser
			} else if errors.Is(err, whatsapp.ErrNoAvailableClient) {
				err = ErrNoAvailableWhatsappClient
			}
		}
	}()

	ctx, err := s.setupTemplateContext(msg, opts)
	if err != nil {
		return
	}

	language, err := s.WhatsappService.ResolveOTPTemplateLanguage()
	if err != nil {
		return
	}

	data, err := s.Translation.WhatsappMessageData(language, msg.spec, ctx)
	if err != nil {
		return
	}

	prepared, err := s.WhatsappService.PrepareOTPTemplate(language, data.Body, opts.OTP)
	if err != nil {
		return
	}

	msg.whatsapp.Options.TemplateName = prepared.TemplateName
	msg.whatsapp.Options.TemplateType = prepared.TemplateType
	msg.whatsapp.Options.Language = prepared.Language
	msg.whatsapp.Options.Components = prepared.Components
	msg.whatsapp.Options.Namespace = prepared.Namespace

	err = msg.whatsapp.Send(s.WhatsappService)
	if err != nil {
		return
	}

	return
}
