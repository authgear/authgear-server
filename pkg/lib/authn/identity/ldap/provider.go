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

func (p *Provider) List(userID string) ([]*identity.LDAP, error) {
	return p.Store.List(userID)
}

func (p *Provider) GetByServerUserID(serverName string, userIDAttribute string, userIDAttributeValue string) (*identity.LDAP, error) {
	return p.Store.GetByServerUserID(serverName, userIDAttribute, userIDAttributeValue)
}
