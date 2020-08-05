package oob

import (
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/uuid"
	"github.com/authgear/authgear-server/pkg/otp"
)

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type OTPMessageSender interface {
	SendEmail(email string, opts otp.SendOptions, message config.EmailMessageConfig) error
	SendSMS(phone string, opts otp.SendOptions, message config.SMSMessageConfig) error
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

func (p *Provider) New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string, tag []string) *Authenticator {
	if tag == nil {
		tag = []string{}
	}
	a := &Authenticator{
		ID:      uuid.New(),
		UserID:  userID,
		Channel: channel,
		Phone:   phone,
		Email:   email,
		Tag:     tag,
	}
	return a
}

func (p *Provider) Create(a *Authenticator) error {
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
	target string,
	code string,
	messageType otp.MessageType,
) (result *otp.OOBSendResult, err error) {
	opts := otp.SendOptions{
		OTP:         code,
		MessageType: messageType,
	}
	switch channel {
	case authn.AuthenticatorOOBChannelEmail:
		err = p.OTPMessageSender.SendEmail(target, opts, p.Config.Email.Message)
	case authn.AuthenticatorOOBChannelSMS:
		err = p.OTPMessageSender.SendSMS(target, opts, p.Config.SMS.Message)
	default:
		panic("oob: unknown channel type: " + channel)
	}

	if err != nil {
		return
	}

	result = &otp.OOBSendResult{
		Channel:      string(channel),
		CodeLength:   len(code),
		SendCooldown: OOBOTPSendCooldownSeconds,
	}
	return
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
