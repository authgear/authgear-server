package whatsapp

import (
	"context"
	"fmt"
	"strings"

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

func (s *Service) makeAuthenticationTemplateComponents(text string, code string) ([]TemplateComponent, error) {

	var component []TemplateComponent = []TemplateComponent{}

	// The text cannot include any newline characters
	text = strings.ReplaceAll(text, "\n", "")

	body := NewTemplateComponent(TemplateComponentTypeBody)
	bodyParam := NewTemplateComponentTextParameter(text)
	body.Parameters = append(body.Parameters, *bodyParam)
	component = append(component, *body)

	button := NewTemplateButtonComponent(TemplateComponentSubTypeURL, 0)
	buttonParam := NewTemplateComponentTextParameter(code)
	button.Parameters = append(button.Parameters, *buttonParam)
	component = append(component, *button)

	return component, nil
}

func (s *Service) getOTPTemplate() (*config.WhatsappTemplateConfig, error) {
	switch s.WhatsappConfig.APIType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return nil, ErrNoAvailableClient
		}
		return s.OnPremisesClient.GetOTPTemplate(), nil
	default:
		return nil, fmt.Errorf("whatsapp: unknown api type")
	}
}

func (s *Service) ResolveOTPTemplateLanguage(ctx context.Context) (lang string, err error) {
	template, err := s.getOTPTemplate()
	if err != nil {
		return "", err
	}
	lang = s.resolveTemplateLanguage(ctx, template.Languages)
	return
}

func (s *Service) PrepareOTPTemplate(language string, text string, code string) (*PreparedOTPTemplate, error) {
	template, err := s.getOTPTemplate()
	if err != nil {
		return nil, err
	}

	var component []TemplateComponent = []TemplateComponent{}

	switch template.Type {
	case config.WhatsappTemplateTypeAuthentication:
		c, err := s.makeAuthenticationTemplateComponents(text, code)
		if err != nil {
			return nil, err
		}
		component = c
	default:
		panic("whatsapp: unknown template type")
	}

	return &PreparedOTPTemplate{
		TemplateName: template.Name,
		TemplateType: string(template.Type),
		Language:     language,
		Components:   component,
		Namespace:    template.Namespace,
	}, nil
}

func (s *Service) SendTemplate(ctx context.Context, opts *SendTemplateOptions) error {
	switch s.WhatsappConfig.APIType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return ErrNoAvailableClient
		}
		return s.OnPremisesClient.SendTemplate(
			ctx,
			opts.To,
			opts.TemplateName,
			opts.Language,
			opts.Components,
			opts.Namespace)
	default:
		return fmt.Errorf("whatsapp: unknown api type")
	}
}
