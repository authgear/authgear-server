package totp

import (
	"sort"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store  *Store
	Config *config.AuthenticatorTOTPConfiguration
	Time   time.Provider
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

func (p *Provider) Create(userID string, displayName string) (*Authenticator, error) {
	secret, err := GenerateSecret()
	if err != nil {
		return nil, err
	}

	a := &Authenticator{
		ID:          uuid.New(),
		UserID:      userID,
		CreatedAt:   p.Time.NowUTC(),
		Secret:      secret,
		DisplayName: displayName,
	}

	err = p.Store.Create(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
