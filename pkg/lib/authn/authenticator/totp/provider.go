package totp

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store  *Store
	Config *config.AuthenticatorTOTPConfig
	Clock  clock.Clock
}

func (p *Provider) Get(userID string, id string) (*authenticator.TOTP, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*authenticator.TOTP, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *authenticator.TOTP) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*authenticator.TOTP, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(id string, userID string, totpSpec *authenticator.TOTPSpec, isDefault bool, kind string) (*authenticator.TOTP, error) {
	if id == "" {
		id = uuid.New()
	}

	var secret string
	switch {
	case totpSpec.Secret != "":
		totp, err := secretcode.NewTOTPFromSecret(totpSpec.Secret)
		if err != nil {
			return nil, TranslateTOTPError(err)
		}
		secret = totp.Secret
	default:
		totp, err := secretcode.NewTOTPFromRNG()
		if err != nil {
			return nil, err
		}
		secret = totp.Secret
	}

	a := &authenticator.TOTP{
		ID:          id,
		UserID:      userID,
		Secret:      secret,
		DisplayName: totpSpec.DisplayName,
		IsDefault:   isDefault,
		Kind:        kind,
	}
	return a, nil
}

func (p *Provider) Create(a *authenticator.TOTP) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(a)
}

func (p *Provider) Authenticate(a *authenticator.TOTP, code string) error {
	now := p.Clock.NowUTC()
	totp, err := secretcode.NewTOTPFromSecret(a.Secret)
	if err != nil {
		return err
	}
	if totp.ValidateCode(now, code) {
		return nil
	}

	return ErrInvalidCode
}

func sortAuthenticators(as []*authenticator.TOTP) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
