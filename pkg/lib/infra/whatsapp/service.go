package whatsapp

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type Service struct {
	WhatsappConfig        *config.WhatsappConfig
	LocalizationConfig    *config.LocalizationConfig
	GlobalWhatsappAPIType config.GlobalWhatsappAPIType
	OnPremisesClient      *OnPremisesClient
	CloudAPIClient        *CloudAPIClient
	MessageStore          *MessageStore
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

func (s *Service) SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions) error {
	switch s.WhatsappConfig.GetAPIType(s.GlobalWhatsappAPIType) {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return ErrNoAvailableWhatsappClient
		}

		otpTemplate := s.OnPremisesClient.GetOTPTemplate()
		lang := s.resolveTemplateLanguage(ctx, otpTemplate.Languages)
		components := s.prepareOTPComponents(otpTemplate, opts.OTP)

		return s.OnPremisesClient.SendTemplate(
			ctx,
			opts.To,
			otpTemplate,
			lang,
			components)
	case config.WhatsappAPITypeCloudAPI:
		if s.CloudAPIClient == nil {
			return ErrNoAvailableWhatsappClient
		}

		configuredLanguages := s.CloudAPIClient.GetLanguages()
		lang := s.resolveTemplateLanguage(ctx, configuredLanguages)
		return s.CloudAPIClient.SendAuthenticationOTP(
			ctx,
			opts,
			lang,
		)
	default:
		panic(fmt.Errorf("whatsapp: unknown api type"))
	}
}

func (s *Service) UpdateMessageStatus(ctx context.Context, messageID string, status WhatsappMessageStatus) error {
	return s.MessageStore.UpdateMessageStatus(ctx, messageID, status)
}
