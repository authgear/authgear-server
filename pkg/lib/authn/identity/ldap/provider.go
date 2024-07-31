package ldap

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
}

func (p *Provider) Get(userID string, id string) (*identity.LDAP, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*identity.LDAP, error) {
	return p.Store.GetMany(ids)
}
