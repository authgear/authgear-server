package totp

import (
	"fmt"
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

func (p *Provider) New(userID string, displayName string) *Authenticator {
	secret, err := GenerateSecret()
	if err != nil {
		panic(fmt.Errorf("totp: failed to generate secret: %w", err))
	}

	a := &Authenticator{
		ID:          uuid.New(),
		UserID:      userID,
		Secret:      secret,
		DisplayName: displayName,
	}
	return a
}

func (p *Provider) Create(a *Authenticator) error {
	now := p.Time.NowUTC()
	a.CreatedAt = now

	return p.Store.Create(a)
}

func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
