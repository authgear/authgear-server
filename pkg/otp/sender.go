package otp

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	taskspec "github.com/authgear/authgear-server/pkg/auth/task/spec"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/mail"
	"github.com/authgear/authgear-server/pkg/sms"
	"github.com/authgear/authgear-server/pkg/task"
	"github.com/authgear/authgear-server/pkg/template"
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
	LoginIDType config.LoginIDKeyType
	LoginID     *loginid.LoginID
	OTP         string
	Operation   OOBOperationType
	Stage       OOBAuthenticationStage
}

func (s *MessageSender) makeContext(opts SendOptions) *MessageTemplateContext {
	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	appName := intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(s.Localization.FallbackLanguage), s.AppMetadata, "app_name")

	ctx := &MessageTemplateContext{
		AppName: appName,
		// To be filled by caller
		Email:                "",
		Phone:                "",
		LoginID:              opts.LoginID,
		Code:                 opts.OTP,
		Host:                 s.Endpoints.BaseURL().Host,
		Operation:            opts.Operation,
		Stage:                opts.Stage,
		StaticAssetURLPrefix: s.ServerConfig.StaticAsset.URLPrefix,
	}

	return ctx
}

func (s *MessageSender) SendEmail(opts SendOptions, message config.EmailMessageConfig) (err error) {
	ctx := s.makeContext(opts)
	ctx.Email = opts.LoginID.Value

	textBody, err := s.TemplateEngine.RenderTemplate(
		TemplateItemTypeOTPMessageEmailTXT,
		ctx,
	)
	if err != nil {
		return
	}

	htmlBody, err := s.TemplateEngine.RenderTemplate(
		TemplateItemTypeOTPMessageEmailHTML,
		ctx,
	)
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

func (s *MessageSender) SendSMS(opts SendOptions, message config.SMSMessageConfig) (err error) {
	ctx := s.makeContext(opts)
	ctx.Phone = opts.LoginID.Value

	body, err := s.TemplateEngine.RenderTemplate(
		TemplateItemTypeOTPMessageSMSTXT,
		ctx,
	)
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
