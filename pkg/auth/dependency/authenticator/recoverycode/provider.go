package recoverycode

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/uuid"
)

type Provider struct {
	Store  *Store
	Config *config.AuthenticatorRecoveryCodeConfig
	Clock  clock.Clock
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) List(userID string) ([]*Authenticator, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) Generate(userID string) []*Authenticator {
	var authenticators []*Authenticator
	for i := 0; i < p.Config.Count; i++ {
		a := &Authenticator{
			ID:       uuid.New(),
			UserID:   userID,
			Code:     GenerateCode(),
			Consumed: false,
		}
		authenticators = append(authenticators, a)
	}

	sortAuthenticators(authenticators)
	return authenticators
}

func (p *Provider) ReplaceAll(userID string, as []*Authenticator) error {
	now := p.Clock.NowUTC()
	for _, a := range as {
		a.CreatedAt = now
	}

	err := p.Store.DeleteAll(userID)
	if err != nil {
		return err
	}

	err = p.Store.CreateAll(as)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) Authenticate(candidates []*Authenticator, code string) *Authenticator {
	for _, a := range candidates {
		if VerifyCode(a.Code, code) {
			return a
		}
	}
	return nil
}

func (p *Provider) Consume(authenticator *Authenticator) error {
	return p.Store.MarkConsumed(authenticator)
}

func sortAuthenticators(authenticators []*Authenticator) {
	sort.Slice(authenticators, func(i, j int) bool {
		return authenticators[i].Code < authenticators[j].Code
	})
}
