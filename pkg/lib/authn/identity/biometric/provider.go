package biometric

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
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

func (p *Provider) GetByKeyID(keyID string) (*Identity, error) {
	return p.Store.GetByKeyID(keyID)
}

func (p *Provider) GetMany(ids []string) ([]*Identity, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) New(
	userID string,
	keyID string,
	key []byte,
	deviceInfo map[string]interface{},
) *Identity {
	i := &Identity{
		ID:         uuid.New(),
		Labels:     make(map[string]interface{}),
		UserID:     userID,
		KeyID:      keyID,
		Key:        key,
		DeviceInfo: deviceInfo,
	}
	return i
}

func (p *Provider) Create(i *Identity) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Delete(i *Identity) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
