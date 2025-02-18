package otp

import (
	"context"
	neturl "net/url"
	"path/filepath"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type AdditionalContext struct {
	HasPassword bool
}

type SendOptions struct {
	Channel                 model.AuthenticatorOOBChannel
	Target                  string
	Form                    Form
	Type                    translation.MessageType
	OTP                     string
	AdditionalContext       *AdditionalContext
	IsAdminAPIResetPassword bool
}

type EndpointsProvider interface {
	Origin() *neturl.URL
	LoginLinkVerificationEndpointURL() *neturl.URL
	ResetPasswordEndpointURL() *neturl.URL
}

type TranslationService interface {
	EmailMessageData(ctx context.Context, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.EmailMessageData, error)
	SMSMessageData(ctx context.Context, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.SMSMessageData, error)
	WhatsappMessageData(ctx context.Context, language string, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.WhatsappMessageData, error)
}

type Sender interface {
	SendEmailInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error
	SendSMSImmediately(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error
	SendWhatsappImmediately(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendAuthenticationOTPOptions) error
}

type MessageSender struct {
	AppID       config.AppID
	Translation TranslationService
	Endpoints   EndpointsProvider
	Sender      Sender
}

var FromAdminAPIQueryKey = "x_from_admin_api"

func (s *MessageSender) setupTemplateContext(msgType translation.MessageType, opts SendOptions) (*translation.PartialTemplateVariables, error) {
	url := ""
	if opts.Form == FormLink {
		var linkURL *neturl.URL
		switch msgType {
		case translation.MessageTypeSetupPrimaryOOB,
			translation.MessageTypeSetupSecondaryOOB,
			translation.MessageTypeAuthenticatePrimaryOOB,
			translation.MessageTypeAuthenticateSecondaryOOB:

			linkURL = s.Endpoints.LoginLinkVerificationEndpointURL()
			query := linkURL.Query()
			query.Set("code", opts.OTP)
			linkURL.RawQuery = query.Encode()

		case translation.MessageTypeForgotPassword:

			linkURL = s.Endpoints.ResetPasswordEndpointURL()
			query := linkURL.Query()
			query.Set("code", opts.OTP)
			if opts.IsAdminAPIResetPassword {
				query.Set(FromAdminAPIQueryKey, "true")
			}
			linkURL.RawQuery = query.Encode()

		default:
			panic("otp: unexpected message type for link: " + msgType)
		}

		url = linkURL.String()
	}

	ctx := &translation.PartialTemplateVariables{
		Code: opts.OTP,
		URL:  url,
		Link: url,
		Host: s.Endpoints.Origin().Host,
	}

	switch opts.Channel {
	case model.AuthenticatorOOBChannelEmail:
		ctx.Email = opts.Target
	case model.AuthenticatorOOBChannelSMS:
		ctx.Phone = opts.Target
	case model.AuthenticatorOOBChannelWhatsapp:
		ctx.Phone = opts.Target
	default:
		panic("otp: unknown channel: " + opts.Channel)
	}

	if opts.AdditionalContext != nil {
		ctx.HasPassword = opts.AdditionalContext.HasPassword
	}

	return ctx, nil
}

func (s *MessageSender) selectMessage(form Form, typ translation.MessageType) *translation.MessageSpec {
	var spec *translation.MessageSpec
	switch typ {
	case translation.MessageTypeVerification:
		spec = translation.MessageVerification
	case translation.MessageTypeSetupPrimaryOOB:
		if form == FormLink {
			spec = translation.MessageSetupPrimaryLoginLink
		} else {
			spec = translation.MessageSetupPrimaryOOB
		}
	case translation.MessageTypeSetupSecondaryOOB:
		if form == FormLink {
			spec = translation.MessageSetupSecondaryLoginLink
		} else {
			spec = translation.MessageSetupSecondaryOOB
		}
	case translation.MessageTypeAuthenticatePrimaryOOB:
		if form == FormLink {
			spec = translation.MessageAuthenticatePrimaryLoginLink
		} else {
			spec = translation.MessageAuthenticatePrimaryOOB
		}
	case translation.MessageTypeAuthenticateSecondaryOOB:
		if form == FormLink {
			spec = translation.MessageAuthenticateSecondaryLoginLink
		} else {
			spec = translation.MessageAuthenticateSecondaryOOB
		}
	case translation.MessageTypeForgotPassword:
		if form == FormLink {
			spec = translation.MessageForgotPasswordLink
		} else {
			spec = translation.MessageForgotPasswordOOB
		}
	case translation.MessageTypeWhatsappCode:
		spec = translation.MessageWhatsappCode
	default:
		panic("otp: unknown message type: " + typ)
	}

	return spec
}

func (s *MessageSender) sendEmail(ctx context.Context, opts SendOptions) error {
	spec := s.selectMessage(opts.Form, opts.Type)
	msgType := spec.MessageType

	variables, err := s.setupTemplateContext(msgType, opts)
	if err != nil {
		return err
	}

	data, err := s.Translation.EmailMessageData(ctx, spec, variables)
	if err != nil {
		return err
	}

	mailSendOptions := &mail.SendOptions{
		Sender:    data.Sender,
		ReplyTo:   data.ReplyTo,
		Subject:   data.Subject,
		Recipient: opts.Target,
		TextBody:  data.TextBody.String,
		HTMLBody:  data.HTMLBody.String,
	}

	err = s.Sender.SendEmailInNewGoroutine(ctx, msgType, mailSendOptions)
	if err != nil {
		return err
	}

	return nil
}

func (s *MessageSender) sendSMS(ctx context.Context, opts SendOptions) error {
	spec := s.selectMessage(opts.Form, opts.Type)
	msgType := spec.MessageType

	variables, err := s.setupTemplateContext(msgType, opts)
	if err != nil {
		return err
	}

	data, err := s.Translation.SMSMessageData(ctx, spec, variables)
	if err != nil {
		return err
	}

	smsSendOptions := &sms.SendOptions{
		Sender:            data.Sender,
		To:                opts.Target,
		Body:              data.Body.String,
		AppID:             string(s.AppID),
		TemplateName:      filepath.Base(spec.SMSTemplate.Name),
		LanguageTag:       data.Body.LanguageTag,
		TemplateVariables: sms.NewTemplateVariablesFromPreparedTemplateVariables(data.PreparedTemplateVariables),
	}

	err = s.Sender.SendSMSImmediately(ctx, msgType, smsSendOptions)
	if err != nil {
		return err
	}

	return nil
}

func (s *MessageSender) sendWhatsapp(ctx context.Context, opts SendOptions) (err error) {

	spec := s.selectMessage(opts.Form, opts.Type)
	msgType := spec.MessageType

	whatsappSendAuthenticationOTPOptions := &whatsapp.SendAuthenticationOTPOptions{
		To:  opts.Target,
		OTP: opts.OTP,
	}

	err = s.Sender.SendWhatsappImmediately(ctx, msgType, whatsappSendAuthenticationOTPOptions)
	if err != nil {
		return
	}

	return
}

func (s *MessageSender) Send(ctx context.Context, opts SendOptions) error {
	switch opts.Channel {
	case model.AuthenticatorOOBChannelEmail:
		err := s.sendEmail(ctx, opts)
		if err != nil {
			return err
		}

		return nil
	case model.AuthenticatorOOBChannelSMS:
		err := s.sendSMS(ctx, opts)
		if err != nil {
			return err
		}

		return nil
	case model.AuthenticatorOOBChannelWhatsapp:
		err := s.sendWhatsapp(ctx, opts)
		if err != nil {
			return err
		}

		return nil
	default:
		panic("otp: unknown channel: " + opts.Channel)
	}
}
