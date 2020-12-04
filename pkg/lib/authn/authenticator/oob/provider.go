package oob

import (
	"errors"
	"sort"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type CodeStore interface {
	Create(code *Code) error
	Get(authenticatorID string) (*Code, error)
	Delete(authenticatorID string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("oob_otp")} }

type Provider struct {
	Config    *config.AuthenticatorOOBConfig
	Store     *Store
	CodeStore CodeStore
	Clock     clock.Clock
	Logger    Logger
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*Authenticator, error) {
	return p.Store.GetMany(ids)
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

func (p *Provider) New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string, isDefault bool, kind string) *Authenticator {
	a := &Authenticator{
		ID:        uuid.New(),
		Labels:    make(map[string]interface{}),
		UserID:    userID,
		Channel:   channel,
		Phone:     phone,
		Email:     email,
		IsDefault: isDefault,
		Kind:      kind,
	}
	return a
}

func (p *Provider) Create(a *Authenticator) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
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

func (p *Provider) GetCode(authenticatorID string) (*Code, error) {
	return p.CodeStore.Get(authenticatorID)
}

func (p *Provider) CreateCode(authenticatorID string) (*Code, error) {
	code := otp.FormatNumeric.Generate()
	codeModel := &Code{
		AuthenticatorID: authenticatorID,
		Code:            code,
		// TODO(oob): Expiry should be configurable
		ExpireAt: p.Clock.NowUTC().Add(time.Duration(3600) * time.Second),
	}

	err := p.CodeStore.Create(codeModel)
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (p *Provider) VerifyCode(authenticatorID string, code string) (*Code, error) {
	codeModel, err := p.CodeStore.Get(authenticatorID)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidCode
	} else if err != nil {
		return nil, err
	}

	if !otp.ValidateOTP(code, codeModel.Code) {
		return nil, ErrInvalidCode
	}

	if err = p.CodeStore.Delete(authenticatorID); err != nil {
		p.Logger.WithError(err).Error("failed to delete code after validation")
	}

	return codeModel, nil
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
