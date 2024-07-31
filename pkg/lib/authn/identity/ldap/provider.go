package ldap

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var _ service.LDAPIdentityProvider = &Provider{}

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

func (p *Provider) WithUpdate(iden *identity.LDAP, claims map[string]interface{}, rawEntryJSON map[string]interface{}) *identity.LDAP {
	newIden := *iden
	newIden.Claims = claims
	newIden.RawEntryJSON = rawEntryJSON
	return &newIden
}

func (p *Provider) Create(i *identity.LDAP) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Update(i *identity.LDAP) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(i)
}

func (p *Provider) Delete(i *identity.LDAP) error {
	return p.Store.Delete(i)
}
