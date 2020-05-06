package anonymous

import (
	"sort"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store *Store
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

func (p *Provider) GetByKeyID(keyID string) (*Identity, error) {
	return p.Store.GetByKeyID(keyID)
}

func (p *Provider) New(
	userID string,
	keyID string,
	key string,
) *Identity {
	i := &Identity{
		ID:     uuid.New(),
		UserID: userID,
		KeyID:  keyID,
		Key:    key,
	}
	return i
}

func (p *Provider) Create(i *Identity) error {
	return p.Store.Create(i)
}

func (p *Provider) Delete(i *Identity) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].KeyID < is[j].KeyID
	})
}
