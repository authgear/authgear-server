package recoverycode

import (
	"errors"
	"sort"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store  *Store
	Config *config.AuthenticatorRecoveryCodeConfiguration
	Time   time.Provider
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
	now := p.Time.NowUTC()
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

func (p *Provider) Authenticate(authenticator *Authenticator, code string) error {
	ok := VerifyCode(authenticator.Code, code)
	if !ok {
		return errors.New("invalid recovery code")
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
