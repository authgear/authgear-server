package ldap

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ldap"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var _ service.LDAPIdentityProvider = &Provider{}

type StandardAttributesNormalizer interface {
	Normalize(stdattrs.T) error
}

type Provider struct {
	Store                        *Store
	Clock                        clock.Clock
	StandardAttributesNormalizer StandardAttributesNormalizer
}

func (p *Provider) Get(userID string, id string) (*identity.LDAP, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetMany(ids []string) ([]*identity.LDAP, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) List(userID string) ([]*identity.LDAP, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}
	sortIdentities(is)
	return is, nil
}

func (p *Provider) GetByServerUserID(serverName string, userIDAttributeName string, userIDAttributeValue []byte) (*identity.LDAP, error) {
	return p.Store.GetByServerUserID(serverName, userIDAttributeName, userIDAttributeValue)
}

func (p *Provider) ListByClaim(name string, value string) ([]*identity.LDAP, error) {
	is, err := p.Store.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	sortIdentities(is)
	return is, nil
}

func (p *Provider) New(
	userID string,
	serverName string,
	userIDAttributeName string,
	userIDAttributeValue []byte,
	claims map[string]interface{},
	rawEntryJSON map[string]interface{},
) *identity.LDAP {
	if claims == nil {
		claims = make(map[string]interface{})
	}
	if rawEntryJSON == nil {
		rawEntryJSON = make(map[string]interface{})
	}
	return &identity.LDAP{
		ID:                   uuid.New(),
		UserID:               userID,
		ServerName:           serverName,
		UserIDAttributeName:  userIDAttributeName,
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

func sortIdentities(is []*identity.LDAP) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}

func (p *Provider) CreateNormalizedIdentitySpecFromLDAPEntry(serverConfig *config.LDAPServerConfig, entry *ldap.Entry) (*identity.Spec, error) {
	userIDAttributeName := serverConfig.UserIDAttributeName
	userIDAttributeValue := entry.GetRawAttributeValue(userIDAttributeName)

	claims := map[string]interface{}{}
	if v := entry.GetAttributeValue(ldap.AttributeMail.Name); v != "" {
		claims[string(model.ClaimEmail)] = v
	}
	if v := entry.GetAttributeValue(ldap.AttributeMobile.Name); v != "" {
		claims[string(model.ClaimPhoneNumber)] = v
	}
	if v := entry.GetAttributeValue(ldap.AttributeUID.Name); v != "" {
		claims[string(model.ClaimPreferredUsername)] = v
	}

	err := p.StandardAttributesNormalizer.Normalize(claims)
	if err != nil {
		claims = map[string]interface{}{}
	}

	return &identity.Spec{
		Type: model.IdentityTypeLDAP,
		LDAP: &identity.LDAPSpec{
			ServerName:           serverConfig.Name,
			UserIDAttributeName:  userIDAttributeName,
			UserIDAttributeValue: userIDAttributeValue,
			Claims:               claims,
			RawEntryJSON:         entry.ToJSON(),
		},
	}, nil
}
