package whatsapp

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("whatsapp-service")}
}

type Service struct {
	Logger             ServiceLogger
	WhatsappConfig     *config.WhatsappConfig
	LocalizationConfig *config.LocalizationConfig
	OnPremisesClient   *OnPremisesClient
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

func (s *Service) makeAuthenticationTemplateComponents(code string) []TemplateComponent {
	// See https://developers.facebook.com/docs/whatsapp/api/messages/message-templates/authentication-message-templates

	var component []TemplateComponent = []TemplateComponent{}

	body := NewTemplateComponent(TemplateComponentTypeBody)
	// The body is just the code.
	bodyParam := NewTemplateComponentTextParameter(code)
	body.Parameters = append(body.Parameters, *bodyParam)
	component = append(component, *body)

	button := NewTemplateButtonComponent(TemplateComponentSubTypeURL, 0)
	// The button copies the code.
	buttonParam := NewTemplateComponentTextParameter(code)
	button.Parameters = append(button.Parameters, *buttonParam)
	component = append(component, *button)

	return component
}

func (s *Service) prepareOTPComponents(template *config.WhatsappTemplateConfig, code string) []TemplateComponent {
	switch template.Type {
	case config.WhatsappTemplateTypeAuthentication:
		return s.makeAuthenticationTemplateComponents(code)
	default:
		panic("whatsapp: unknown template type")
	}
}

func (s *Service) ResolveSendAuthenticationOTPOptions(ctx context.Context, opts *SendAuthenticationOTPOptions) (*ResolvedSendAuthenticationOTPOptions, error) {
	switch s.WhatsappConfig.APIType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return nil, ErrNoAvailableWhatsappClient
		}

		otpTemplate := s.OnPremisesClient.GetOTPTemplate()
		lang := s.resolveTemplateLanguage(ctx, otpTemplate.Languages)
		components := s.prepareOTPComponents(otpTemplate, opts.OTP)

		return &ResolvedSendAuthenticationOTPOptions{
			To:                 opts.To,
			OTP:                opts.OTP,
			TemplateName:       otpTemplate.Name,
			TemplateLanguage:   lang,
			TemplateNamespace:  otpTemplate.Namespace,
			TemplateComponents: components,
		}, nil
	default:
		panic(fmt.Errorf("whatsapp: unknown api type"))
	}
}

func (s *Service) SendAuthenticationOTP(ctx context.Context, opts *ResolvedSendAuthenticationOTPOptions) error {
	switch s.WhatsappConfig.APIType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return ErrNoAvailableWhatsappClient
		}

		return s.OnPremisesClient.SendTemplate(
			ctx,
			opts.To,
			opts.TemplateName,
			opts.TemplateLanguage,
			opts.TemplateComponents,
			opts.TemplateNamespace)
	default:
		panic(fmt.Errorf("whatsapp: unknown api type"))
	}
}
