package whatsapp

import (
	"context"
	"strings"

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
	InspectState(kind otp.Kind, target string) (*otp.State, error)
	InspectCode(purpose otp.Purpose, target string) (*otp.Code, error)
}

type WhatsappSender interface {
	SendTemplate(
		to string,
		templateName string,
		templateLanguage string,
		templateComponents []whatsapp.TemplateComponent,
		namespace string,
	) error
}

type Provider struct {
	Context        context.Context
	Config         *config.AppConfig
	Events         EventService
	OTPCodeService OTPCodeService
	WhatsappSender WhatsappSender
	WhatsappConfig *config.WhatsappConfig
	TemplateEngine *template.Engine
}

type templateData struct {
	Code string
}

type WhatsappCode struct {
	Code       string
	CodeLength int
	IsNew      bool
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

func (p *Provider) getOTPKind() otp.Kind {
	return otp.KindWhatsapp(p.Config)
}

func (p *Provider) getOTPForm() otp.Form {
	return otp.FormCode
}

func (p *Provider) InspectCodeState(phone string) (*otp.State, error) {
	kind := p.getOTPKind()
	return p.OTPCodeService.InspectState(kind, phone)
}

func (p *Provider) GenerateCode(phone string, webSessionID string, useExistingOnRateLimited bool) (*WhatsappCode, error) {
	kind := p.getOTPKind()
	form := p.getOTPForm()
	code, err := p.OTPCodeService.GenerateOTP(
		kind,
		phone,
		form,
		&otp.GenerateOptions{WebSessionID: webSessionID},
	)
	if apierrors.IsKind(err, ratelimit.RateLimited) && useExistingOnRateLimited {
		// Ignore rate limits; return current OTP
		code, serr := p.OTPCodeService.InspectCode(kind.Purpose(), phone)
		if apierrors.IsKind(serr, otp.InvalidOTPCode) {
			// Current OTP is invalidated; return original rate limit error
			return nil, err
		} else if serr != nil {
			return nil, serr
		}
		return &WhatsappCode{
			Code:       code.Code,
			CodeLength: form.CodeLength(),
			IsNew:      false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &WhatsappCode{
		Code:       code,
		CodeLength: form.CodeLength(),
		IsNew:      true,
	}, nil
}

func (p *Provider) makeAuthenticationTemplateComponent(code string, language string) ([]whatsapp.TemplateComponent, error) {

	var component []whatsapp.TemplateComponent = []whatsapp.TemplateComponent{}

	data := make(map[string]any)
	template.Embed(data, templateData{
		Code: code,
	})

	text, err := p.TemplateEngine.Render(otp.TemplateWhatsappOTPCodeTXT, []string{language}, data)
	if err != nil {
		return nil, err
	}
	// The text cannot include any newline characters
	text = strings.ReplaceAll(text, "\n", "")

	body := whatsapp.NewTemplateComponent(whatsapp.TemplateComponentTypeBody)
	bodyParam := whatsapp.NewTemplateComponentTextParameter(text)
	body.Parameters = append(body.Parameters, *bodyParam)
	component = append(component, *body)

	button := whatsapp.NewTemplateButtonComponent(whatsapp.TemplateComponentSubTypeURL, 0)
	buttonParam := whatsapp.NewTemplateComponentTextParameter(code)
	button.Parameters = append(button.Parameters, *buttonParam)
	component = append(component, *button)

	return component, nil
}

func (p *Provider) SendCode(phone string, code string) error {
	var component []whatsapp.TemplateComponent = []whatsapp.TemplateComponent{}
	template := p.WhatsappConfig.Templates.OTP
	language := p.resolveTemplateLanguage(template.Languages)

	switch template.Type {
	case config.WhatsappTemplateTypeAuthentication:
		c, err := p.makeAuthenticationTemplateComponent(code, language)
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
		template.Namespace,
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
