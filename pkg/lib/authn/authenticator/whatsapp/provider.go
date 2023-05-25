package whatsapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opts *otp.GenerateOptions) (string, error)
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectWhatsappOTP(kind otp.Kind, target string) (string, error)
}

type WhatsappSender interface {
	SendTemplate(
		to string,
		templateName string,
		templateLanguage string,
		templateComponents []whatsapp.TemplateComponent,
	) error
}

type Provider struct {
	Context         context.Context
	Config          *config.AppConfig
	WATICredentials *config.WATICredentials
	Events          EventService
	OTPCodeService  OTPCodeService
	WhatsappSender  WhatsappSender
	WhatsappConfig  *config.WhatsappConfig
	TemplateEngine  *template.Engine
}

type templateData struct {
	Code string
}

func (p *Provider) resolveTemplateLanguage(supportedLanguages []string) string {
	if len(supportedLanguages) < 1 {
		panic("whatsapp: template has no supported language")
	}
	preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
	supportedLanguageTags := intl.Supported(supportedLanguages, intl.Fallback(supportedLanguages[0]))
	idx, _ := intl.Match(preferredLanguageTags, supportedLanguageTags)
	return supportedLanguageTags[idx]
}

func (p *Provider) GetServerWhatsappPhone() string {
	// return the phone from different config when more whatsapp api provider is supported
	if p.WATICredentials != nil {
		return p.WATICredentials.WhatsappPhoneNumber
	}
	return ""
}

func (p *Provider) GenerateCode(phone string, webSessionID string) (string, error) {
	kind := otp.KindWhatsapp(p.Config)
	code, err := p.OTPCodeService.GenerateOTP(
		kind,
		phone,
		otp.FormCode,
		&otp.GenerateOptions{WebSessionID: webSessionID},
	)
	if apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore rate limits; return current OTP
		code, serr := p.OTPCodeService.InspectWhatsappOTP(kind, phone)
		if apierrors.IsKind(serr, otp.InvalidOTPCode) {
			// Current OTP is invalidated; return original rate limit error
			return "", err
		} else if serr != nil {
			return "", serr
		}
		return code, nil
	} else if err != nil {
		return "", err
	}

	return code, nil
}

func (p *Provider) makeAuthenticationTemplateComponent(code string) []whatsapp.TemplateComponent {

	var component []whatsapp.TemplateComponent = []whatsapp.TemplateComponent{}

	body := whatsapp.NewTemplateComponent(whatsapp.TemplateComponentTypeBody)
	bodyParam := whatsapp.NewTemplateComponentTextParameter(code)
	body.Parameters = append(body.Parameters, *bodyParam)
	component = append(component, *body)

	button := whatsapp.NewTemplateButtonComponent(whatsapp.TemplateComponentSubTypeURL, 0)
	buttonParam := whatsapp.NewTemplateComponentTextParameter(code)
	button.Parameters = append(button.Parameters, *buttonParam)
	component = append(component, *button)

	return component
}

func (p *Provider) makeUtilTemplateComponent(code string, language string) ([]whatsapp.TemplateComponent, error) {
	var component []whatsapp.TemplateComponent = []whatsapp.TemplateComponent{}
	tpl := p.WhatsappConfig.Templates.OTP

	data := make(map[string]any)
	template.Embed(data, templateData{
		Code: code,
	})

	if tpl.Components.Header != nil {
		header := whatsapp.NewTemplateComponent(whatsapp.TemplateComponentTypeHeader)

		for _, param := range tpl.Components.Header.Parameters {
			text, err := p.TemplateEngine.RenderString(param, []string{language}, data)
			if err != nil {
				return nil, err
			}
			paramObj := whatsapp.NewTemplateComponentTextParameter(text)
			header.Parameters = append(header.Parameters, *paramObj)
		}

		component = append(component, *header)
	}

	if tpl.Components.Body != nil {
		body := whatsapp.NewTemplateComponent(whatsapp.TemplateComponentTypeBody)

		for _, param := range tpl.Components.Body.Parameters {
			text, err := p.TemplateEngine.RenderString(param, []string{language}, data)
			if err != nil {
				return nil, err
			}
			paramObj := whatsapp.NewTemplateComponentTextParameter(text)
			body.Parameters = append(body.Parameters, *paramObj)
		}

		component = append(component, *body)
	}

	return component, nil
}

func (p *Provider) SendCode(phone string, code string) error {
	var component []whatsapp.TemplateComponent = []whatsapp.TemplateComponent{}
	template := p.WhatsappConfig.Templates.OTP
	language := p.resolveTemplateLanguage(template.Languages)

	switch template.Type {
	case config.WhatsappTemplateTypeAuthentication:
		component = p.makeAuthenticationTemplateComponent(code)
	case config.WhatsappTemplateTypeUtil:
		c, err := p.makeUtilTemplateComponent(code, language)
		if err != nil {
			return err
		}
		component = c
	default:
		panic("whatsapp: unknown template type")
	}

	return p.WhatsappSender.SendTemplate(
		phone,
		p.WhatsappConfig.Templates.OTP.Name,
		language,
		component,
	)
}

func (p *Provider) VerifyCode(phone string, code string, userID string) error {
	err := p.OTPCodeService.VerifyOTP(
		otp.KindWhatsapp(p.Config),
		phone,
		code,
		&otp.VerifyOptions{
			UserID: userID,
		},
	)
	if err != nil {
		return err
	}

	if err := p.Events.DispatchEvent(&nonblocking.WhatsappOTPVerifiedEventPayload{
		Phone: phone,
	}); err != nil {
		return err
	}

	return nil
}
