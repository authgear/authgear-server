package whatsapp

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type OTPCodeService interface {
	GenerateWhatsappCode(target string, appID string, webSessionID string) (*otp.Code, error)
	VerifyWhatsappCode(target string, consume bool) error
}

type Provider struct {
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

func (p *Provider) GenerateCode(phone string, appID string, webSessionID string) (*otp.Code, error) {
	code, err := p.OTPCodeService.GenerateWhatsappCode(phone, appID, webSessionID)
	if errors.Is(err, otp.ErrInvalidCode) {
		return nil, ErrInvalidCode
	} else if errors.Is(err, otp.ErrInputRequired) {
		return nil, ErrInputRequired
	} else if err != nil {
		return nil, err
	}

	return code, nil
}

func (p *Provider) VerifyCode(phone string, consume bool) error {
	err := p.OTPCodeService.VerifyWhatsappCode(phone, consume)
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
