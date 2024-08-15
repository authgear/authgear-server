package ldap

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
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

func (p *Provider) New(
	userID string,
	serverName string,
	userIDAttribute string,
	userIDAttributeValue string,
	claims map[string]interface{},
	rawEntryJSON map[string]interface{},
) *identity.LDAP {
	return &identity.LDAP{
		ID:                   uuid.New(),
		UserID:               userID,
		ServerName:           serverName,
		UserIDAttribute:      userIDAttribute,
		UserIDAttributeValue: userIDAttributeValue,
		Claims:               claims,
		RawEntryJSON:         rawEntryJSON,
	}
}

func (p *Provider) Create(i *identity.LDAP) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}
