package oob

import (
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/auth/metadata"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/uuid"
	"github.com/authgear/authgear-server/pkg/otp"
)

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type OTPMessageSender interface {
	SendEmail(opts otp.SendOptions, message config.EmailMessageConfig) error
	SendSMS(opts otp.SendOptions, message config.SMSMessageConfig) error
}

type Provider struct {
	Config           *config.AuthenticatorOOBConfig
	Store            *Store
	Clock            clock.Clock
	OTPMessageSender OTPMessageSender
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByChannel(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string) (*Authenticator, error) {
	return p.Store.GetByChannel(userID, channel, phone, email)
}

func (p *Provider) Delete(a *Authenticator) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*Authenticator, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string, identityID *string) *Authenticator {
	a := &Authenticator{
		ID:         uuid.New(),
		UserID:     userID,
		Channel:    channel,
		Phone:      phone,
		Email:      email,
		IdentityID: identityID,
	}
	return a
}

func (p *Provider) Create(a *Authenticator) error {
	_, err := p.Store.GetByChannel(a.UserID, a.Channel, a.Phone, a.Email)
	if err == nil {
		return authenticator.ErrAuthenticatorAlreadyExists
	} else if !errors.Is(err, authenticator.ErrAuthenticatorNotFound) {
		return err
	}

	now := p.Clock.NowUTC()
	a.CreatedAt = now

	return p.Store.Create(a)
}

func (p *Provider) getOTPOpts(channel authn.AuthenticatorOOBChannel) otp.ValidateOpts {
	var digits int
	switch channel {
	case authn.AuthenticatorOOBChannelEmail:
		digits = p.Config.Email.CodeDigits
	case authn.AuthenticatorOOBChannelSMS:
		digits = p.Config.SMS.CodeDigits
	default:
		panic("oob: unknown channel type: " + channel)
	}
	return otp.ValidateOptsOOBTOTP(digits)
}

func (p *Provider) Authenticate(secret string, channel authn.AuthenticatorOOBChannel, code string) error {
	ok := otp.ValidateTOTP(secret, code, p.Clock.NowUTC(), p.getOTPOpts(channel))
	if !ok {
		return errors.New("invalid OOB code")
	}
	return nil
}

func (p *Provider) GenerateCode(secret string, channel authn.AuthenticatorOOBChannel) string {
	code, err := otp.GenerateTOTP(secret, p.Clock.NowUTC(), p.getOTPOpts(channel))
	if err != nil {
		panic(fmt.Sprintf("oob: cannot generate code: %s", err))
	}

	return code
}

func (p *Provider) SendCode(
	channel authn.AuthenticatorOOBChannel,
	loginID *loginid.LoginID,
	code string,
	origin otp.MessageOrigin,
	operation otp.OOBOperationType,
) error {
	opts := otp.SendOptions{
		LoginID:   loginID,
		OTP:       code,
		Origin:    origin,
		Operation: operation,
	}
	switch channel {
	case authn.AuthenticatorOOBChannelEmail:
		opts.LoginIDType = config.LoginIDKeyType(metadata.Email)
		return p.OTPMessageSender.SendEmail(opts, p.Config.Email.Message)
	case authn.AuthenticatorOOBChannelSMS:
		opts.LoginIDType = config.LoginIDKeyType(metadata.Phone)
		return p.OTPMessageSender.SendSMS(opts, p.Config.SMS.Message)
	default:
		panic("oob: unknown channel type: " + channel)
	}
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
