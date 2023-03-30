package oob

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
}

func (p *Provider) Get(userID string, id string) (*authenticator.OOBOTP, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*authenticator.OOBOTP, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) Delete(a *authenticator.OOBOTP) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*authenticator.OOBOTP, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(id string, userID string, oobAuthenticatorType model.AuthenticatorType, target string, isDefault bool, kind string) *authenticator.OOBOTP {
	if id == "" {
		id = uuid.New()
	}
	a := &authenticator.OOBOTP{
		ID:                   id,
		UserID:               userID,
		OOBAuthenticatorType: oobAuthenticatorType,
		IsDefault:            isDefault,
		Kind:                 kind,
	}

	switch oobAuthenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		a.Email = target
	case model.AuthenticatorTypeOOBSMS:
		a.Phone = target
	default:
		panic("oob: incompatible authenticator type:" + oobAuthenticatorType)
	}
	return a
}

func (p *Provider) Create(a *authenticator.OOBOTP) error {
	now := p.Clock.NowUTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	return p.Store.Create(a)
}

func (p *Provider) Update(a *authenticator.OOBOTP) error {
	now := p.Clock.NowUTC()
	a.UpdatedAt = now
	return p.Store.Update(a)
}

func sortAuthenticators(as []*authenticator.OOBOTP) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
