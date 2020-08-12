package otp

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	taskspec "github.com/authgear/authgear-server/pkg/auth/task/spec"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type MessageSender struct {
	Context        context.Context
	ServerConfig   *config.ServerConfig
	Localization   *config.LocalizationConfig
	AppMetadata    config.AppMetadata
	Messaging      *config.MessagingConfig
	TemplateEngine *template.Engine
	Endpoints      EndpointsProvider
	TaskQueue      task.Queue
}

type SendOptions struct {
	OTP         string
	URL         string
	MessageType MessageType
}

func (s *MessageSender) makeContext(opts SendOptions) *MessageTemplateContext {
	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	appName := intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(s.Localization.FallbackLanguage), s.AppMetadata, "app_name")

	ctx := &MessageTemplateContext{
		AppName: appName,
		// To be filled by caller
		Email:                "",
		Phone:                "",
		Code:                 opts.OTP,
		URL:                  opts.URL,
		Host:                 s.Endpoints.BaseURL().Host,
		StaticAssetURLPrefix: s.ServerConfig.StaticAsset.URLPrefix,
	}

	return ctx
}

func (s *MessageSender) SendEmail(email string, opts SendOptions, message config.EmailMessageConfig) (err error) {
	ctx := s.makeContext(opts)
	ctx.Email = email

	var textTemplate, htmlTemplate config.TemplateItemType
	switch opts.MessageType {
	case MessageTypeVerification:
		textTemplate = TemplateItemTypeVerificationEmailTXT
		htmlTemplate = TemplateItemTypeVerificationEmailHTML
	case MessageTypeSetupPrimaryOOB:
		textTemplate = TemplateItemTypeSetupPrimaryOOBEmailTXT
		htmlTemplate = TemplateItemTypeSetupPrimaryOOBEmailHTML
	case MessageTypeSetupSecondaryOOB:
		textTemplate = TemplateItemTypeSetupSecondaryOOBEmailTXT
		htmlTemplate = TemplateItemTypeSetupSecondaryOOBEmailHTML
	case MessageTypeAuthenticatePrimaryOOB:
		textTemplate = TemplateItemTypeAuthenticatePrimaryOOBEmailTXT
		htmlTemplate = TemplateItemTypeAuthenticatePrimaryOOBEmailHTML
	case MessageTypeAuthenticateSecondaryOOB:
		textTemplate = TemplateItemTypeAuthenticateSecondaryOOBEmailTXT
		htmlTemplate = TemplateItemTypeAuthenticateSecondaryOOBEmailHTML
	default:
		panic("otp: unknown message type: " + opts.MessageType)
	}

	textBody, err := s.TemplateEngine.RenderTemplate(textTemplate, ctx)
	if err != nil {
		return
	}

	htmlBody, err := s.TemplateEngine.RenderTemplate(htmlTemplate, ctx)
	if err != nil {
		return
	}

	s.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: []mail.SendOptions{
				{
					MessageConfig: config.NewEmailMessageConfig(
						s.Messaging.DefaultEmailMessage,
						message,
					),
					Recipient: ctx.Email,
					TextBody:  textBody,
					HTMLBody:  htmlBody,
				},
			},
		},
	})

	return
}

func (s *MessageSender) SendSMS(phone string, opts SendOptions, message config.SMSMessageConfig) (err error) {
	ctx := s.makeContext(opts)
	ctx.Phone = phone

	var templateType config.TemplateItemType
	switch opts.MessageType {
	case MessageTypeVerification:
		templateType = TemplateItemTypeVerificationSMSTXT
	case MessageTypeSetupPrimaryOOB:
		templateType = TemplateItemTypeSetupPrimaryOOBSMSTXT
	case MessageTypeSetupSecondaryOOB:
		templateType = TemplateItemTypeSetupSecondaryOOBSMSTXT
	case MessageTypeAuthenticatePrimaryOOB:
		templateType = TemplateItemTypeAuthenticatePrimaryOOBSMSTXT
	case MessageTypeAuthenticateSecondaryOOB:
		templateType = TemplateItemTypeAuthenticateSecondaryOOBSMSTXT
	default:
		panic("otp: unknown message type: " + opts.MessageType)
	}

	body, err := s.TemplateEngine.RenderTemplate(templateType, ctx)
	if err != nil {
		return
	}

	s.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			SMSMessages: []sms.SendOptions{
				{
					MessageConfig: config.NewSMSMessageConfig(
						s.Messaging.DefaultSMSMessage,
						message,
					),
					To:   ctx.Phone,
					Body: body,
				},
			},
		},
	})

	return
}
