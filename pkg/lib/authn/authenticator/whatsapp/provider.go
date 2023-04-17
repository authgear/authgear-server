package whatsapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
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

type Provider struct {
	Config          *config.AppConfig
	WATICredentials *config.WATICredentials
	Events          EventService
	OTPCodeService  OTPCodeService
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
