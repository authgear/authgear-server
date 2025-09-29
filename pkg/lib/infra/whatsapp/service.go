package whatsapp

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("whatsapp-service")

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package whatsapp_test

type ServiceCloudAPIClient interface {
	GetLanguages() []string
	SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions, lang string) (result *CloudAPISendAuthenticationOTPResult, err error)
}

type ServiceMessageStore interface {
	GetMessageStatus(ctx context.Context, messageID string) (WhatsappMessageStatus, error)
	UpdateMessageStatus(ctx context.Context, messageID string, status WhatsappMessageStatus) error
}

type SendAuthenticationOTPResult struct {
	MessageID     string
	MessageStatus WhatsappMessageStatus
}

type Service struct {
	Clock                 clock.Clock
	WhatsappConfig        *config.WhatsappConfig
	LocalizationConfig    *config.LocalizationConfig
	GlobalWhatsappAPIType config.GlobalWhatsappAPIType
	OnPremisesClient      *OnPremisesClient
	CloudAPIClient        ServiceCloudAPIClient
	MessageStore          ServiceMessageStore
	Credentials           *config.WhatsappCloudAPICredentials
}

func (s *Service) resolveTemplateLanguage(ctx context.Context, supportedLanguages []string) string {
	if len(supportedLanguages) < 1 {
		panic("whatsapp: template has no supported language")
	}
	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
	configSupportedLanguageTags := intl.Supported(
		s.LocalizationConfig.SupportedLanguages,
		intl.Fallback(*s.LocalizationConfig.FallbackLanguage),
	)
	// First, resolve once based on supported language in config
	// This is to avoid inconsistency of ui language and whatsapp message language
	_, resolvedTag := intl.BestMatch(preferredLanguageTags, configSupportedLanguageTags)
	supportedLanguageTags := intl.Supported(supportedLanguages, intl.Fallback(supportedLanguages[0]))

	// Then, resolve to a language supported by the whatsapp template
	idx, _ := intl.BestMatch([]string{resolvedTag.String()}, supportedLanguageTags)
	return supportedLanguageTags[idx]
}

func (s *Service) makeAuthenticationTemplateComponents(code string) []onPremisesTemplateComponent {
	// See https://developers.facebook.com/docs/whatsapp/api/messages/message-templates/authentication-message-templates

	var component []onPremisesTemplateComponent = []onPremisesTemplateComponent{}

	body := onPremisesNewTemplateComponent(onPremisesTemplateComponentTypeBody)
	// The body is just the code.
	bodyParam := onPremisesNewTemplateComponentTextParameter(code)
	body.Parameters = append(body.Parameters, *bodyParam)
	component = append(component, *body)

	button := onPremisesNewTemplateButtonComponent(onPremisesTemplateComponentSubTypeURL, 0)
	// The button copies the code.
	buttonParam := onPremisesNewTemplateComponentTextParameter(code)
	button.Parameters = append(button.Parameters, *buttonParam)
	component = append(component, *button)

	return component
}

func (s *Service) prepareOTPComponents(template *config.WhatsappOnPremisesOTPTemplateConfig, code string) []onPremisesTemplateComponent {
	switch template.Type {
	case config.WhatsappOnPremisesTemplateTypeAuthentication:
		return s.makeAuthenticationTemplateComponents(code)
	default:
		panic("whatsapp: unknown template type")
	}
}

func (s *Service) SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions) (*SendAuthenticationOTPResult, error) {
	switch s.WhatsappConfig.GetAPIType(s.GlobalWhatsappAPIType) {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return nil, ErrNoAvailableWhatsappClient
		}

		otpTemplate := s.OnPremisesClient.GetOTPTemplate()
		lang := s.resolveTemplateLanguage(ctx, otpTemplate.Languages)
		components := s.prepareOTPComponents(otpTemplate, opts.OTP)

		err := s.OnPremisesClient.SendTemplate(
			ctx,
			opts.To,
			otpTemplate,
			lang,
			components)
		if err != nil {
			return nil, err
		}
		return &SendAuthenticationOTPResult{
			MessageID: "",
			// We don't know the actual status, so always return devlivered
			MessageStatus: WhatsappMessageStatusDelivered,
		}, nil
	case config.WhatsappAPITypeCloudAPI:
		if s.CloudAPIClient == nil {
			return nil, ErrNoAvailableWhatsappClient
		}

		configuredLanguages := s.CloudAPIClient.GetLanguages()
		lang := s.resolveTemplateLanguage(ctx, configuredLanguages)
		cloudAPISendResult, err := s.CloudAPIClient.SendAuthenticationOTP(
			ctx,
			opts,
			lang,
		)
		if err != nil {
			return nil, err
		}
		result := SendAuthenticationOTPResult(*cloudAPISendResult)

		return &result, nil

	default:
		panic(fmt.Errorf("whatsapp: unknown api type"))
	}
}

func (s *Service) UpdateMessageStatus(ctx context.Context, messageID string, status WhatsappMessageStatus) error {
	return s.MessageStore.UpdateMessageStatus(ctx, messageID, status)
}

func (s *Service) GetMessageStatus(ctx context.Context, messageID string) (WhatsappMessageStatus, error) {
	return s.MessageStore.GetMessageStatus(ctx, messageID)
}
