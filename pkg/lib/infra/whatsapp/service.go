package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/sirupsen/logrus"
)

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("whatsapp-service")}
}

type Service struct {
	Context                    context.Context
	Logger                     ServiceLogger
	DevMode                    config.DevMode
	TestModeWhatsappSuppressed config.TestModeWhatsappSuppressed
	Config                     *config.WhatsappConfig
	OnPremisesClient           *OnPremisesClient
	TokenStore                 *TokenStore
}

func (c *Service) logMessage(
	opts *SendTemplateOptions) *logrus.Entry {
	data, _ := json.MarshalIndent(opts.Components, "", "  ")

	return c.Logger.
		WithField("recipient", opts.To).
		WithField("template_name", opts.TemplateName).
		WithField("language", opts.Language).
		WithField("components", string(data)).
		WithField("namespace", opts.Namespace)
}

func (s *Service) resolveTemplateLanguage(supportedLanguages []string) string {
	if len(supportedLanguages) < 1 {
		panic("whatsapp: template has no supported language")
	}
	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	supportedLanguageTags := intl.Supported(supportedLanguages, intl.Fallback(supportedLanguages[0]))
	idx, _ := intl.Match(preferredLanguageTags, supportedLanguageTags)
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
	switch s.Config.APIType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return nil, ErrNoAvailableClient
		}
		return s.OnPremisesClient.GetOTPTemplate(), nil
	default:
		return nil, fmt.Errorf("whatsapp: unknown api type")
	}
}

func (s *Service) ResolveOTPTemplateLanguage() (lang string, err error) {
	template, err := s.getOTPTemplate()
	if err != nil {
		return "", err
	}
	lang = s.resolveTemplateLanguage(template.Languages)
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

func (s *Service) SendTemplate(opts *SendTemplateOptions) error {

	if s.TestModeWhatsappSuppressed {
		s.logMessage(opts).
			Warn("sending whatsapp is suppressed in test mode")
		return nil
	}

	if s.DevMode {
		s.logMessage(opts).
			Warn("skip sending whatsapp in development mode")
		return nil
	}

	switch s.Config.APIType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return ErrNoAvailableClient
		}
		return s.OnPremisesClient.SendTemplate(
			opts.To,
			opts.TemplateName,
			opts.Language,
			opts.Components,
			opts.Namespace)
	default:
		return fmt.Errorf("whatsapp: unknown api type")
	}
}
