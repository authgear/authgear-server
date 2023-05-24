package whatsapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
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
	Config          *config.AppConfig
	WATICredentials *config.WATICredentials
	Events          EventService
	OTPCodeService  OTPCodeService
	WhatsappSender  WhatsappSender
	WhatsappConfig  *config.WhatsappConfig
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

func (p *Provider) SendCode(phone string, code string) error {
	var component []whatsapp.TemplateComponent = []whatsapp.TemplateComponent{}
	template := p.WhatsappConfig.Templates.OTP

	if template.Components.Header != nil {
		header := whatsapp.NewTemplateComponent(whatsapp.TemplateComponentTypeHeader)

		for _, p := range template.Components.Header.Parameters {
			// FIXME: Format it with template engine
			param := whatsapp.NewTemplateComponentTextParameter(p)
			header.Parameters = append(header.Parameters, *param)
		}

		component = append(component, *header)
	}

	if template.Components.Body != nil {
		body := whatsapp.NewTemplateComponent(whatsapp.TemplateComponentTypeBody)

		for _, p := range template.Components.Body.Parameters {
			// FIXME: Format it with template engine
			text := p
			if p == "{{ .Code }}" {
				text = code
			}
			param := whatsapp.NewTemplateComponentTextParameter(text)
			body.Parameters = append(body.Parameters, *param)
		}

		component = append(component, *body)
	}

	return p.WhatsappSender.SendTemplate(
		phone,
		p.WhatsappConfig.Templates.OTP.Name,
		// FIXME: Select a suitable language from user language
		"en",
		component,
	)
}

func (p *Provider) VerifyCode(phone string, consume bool) error {
	err := p.OTPCodeService.VerifyOTP(
		otp.KindWhatsapp(p.Config),
		phone,
		"",
		&otp.VerifyOptions{SkipConsume: !consume, UseSubmittedCode: true},
	)
	if err != nil {
		return err
	}

	if consume {
		if err := p.Events.DispatchEvent(&nonblocking.WhatsappOTPVerifiedEventPayload{
			Phone: phone,
		}); err != nil {
			return err
		}
	}

	return nil
}
