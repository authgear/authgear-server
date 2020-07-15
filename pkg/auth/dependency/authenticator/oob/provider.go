package oob

import (
	"errors"
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

func (p *Provider) Authenticate(expectedCode string, code string) error {
	ok := otp.ValidateOTP(expectedCode, code)
	if !ok {
		return errors.New("invalid bearer token")
	}
	return nil
}

func (p *Provider) GenerateCode(channel authn.AuthenticatorOOBChannel) string {
	var format *otp.Format
	switch channel {
	case authn.AuthenticatorOOBChannelEmail:
		format = otp.GetFormat(p.Config.Email.CodeFormat)
	case authn.AuthenticatorOOBChannelSMS:
		format = otp.GetFormat(p.Config.SMS.CodeFormat)
	default:
		panic("oob: unknown channel type: " + channel)
	}
	return format.Generate()
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
