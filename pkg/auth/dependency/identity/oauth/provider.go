package oauth

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/uuid"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
}

func (p *Provider) List(userID string) ([]*Identity, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(name string, value string) ([]*Identity, error) {
	is, err := p.Store.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(userID, id string) (*Identity, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByProviderSubject(provider config.ProviderID, subjectID string) (*Identity, error) {
	return p.Store.GetByProviderSubject(provider, subjectID)
}

func (p *Provider) GetByUserProvider(userID string, provider config.ProviderID) (*Identity, error) {
	return p.Store.GetByUserProvider(userID, provider)
}

func (p *Provider) New(
	userID string,
	provider config.ProviderID,
	subjectID string,
	profile map[string]interface{},
	claims map[string]interface{},
) *Identity {
	i := &Identity{
		ID:                uuid.New(),
		UserID:            userID,
		ProviderID:        provider,
		ProviderSubjectID: subjectID,
		UserProfile:       profile,
		Claims:            claims,
	}
	return i
}

func (p *Provider) CheckDuplicated(standardClaims map[string]string, userID string) (*Identity, error) {
	// check duplication with standard claims
	for name, value := range standardClaims {
		ls, err := p.ListByClaim(name, value)
		if err != nil {
			return nil, err
		}

		for _, i := range ls {
			if i.UserID == userID {
				continue
			}
			return i, identity.ErrIdentityAlreadyExists
		}
	}

	return nil, nil
}

func (p *Provider) Create(i *Identity) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Update(i *Identity) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(i)
}

func (p *Provider) Delete(i *Identity) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
