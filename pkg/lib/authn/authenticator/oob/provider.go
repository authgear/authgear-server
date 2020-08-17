package oob

import (
	"errors"
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Config *config.AuthenticatorOOBConfig
	Store  *Store
	Clock  clock.Clock
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

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
