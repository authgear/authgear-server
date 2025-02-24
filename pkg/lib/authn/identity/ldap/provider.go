package ldap

import (
	"context"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ldap"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type StandardAttributesNormalizer interface {
	Normalize(context.Context, stdattrs.T) error
}

type Provider struct {
	Store                        *Store
	Clock                        clock.Clock
	StandardAttributesNormalizer StandardAttributesNormalizer
}

func (p *Provider) Get(ctx context.Context, userID string, id string) (*identity.LDAP, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*identity.LDAP, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) List(ctx context.Context, userID string) ([]*identity.LDAP, error) {
	is, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	sortIdentities(is)
	return is, nil
}

func (p *Provider) GetByServerUserID(ctx context.Context, serverName string, userIDAttributeName string, userIDAttributeValue []byte) (*identity.LDAP, error) {
	return p.Store.GetByServerUserID(ctx, serverName, userIDAttributeName, userIDAttributeValue)
}

func (p *Provider) ListByClaim(ctx context.Context, name string, value string) ([]*identity.LDAP, error) {
	is, err := p.Store.ListByClaim(ctx, name, value)
	if err != nil {
		return nil, err
	}
	sortIdentities(is)
	return is, nil
}

func (p *Provider) New(
	userID string,
	serverName string,
	loginUserName *string,
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
		LastLoginUserName:    loginUserName,
	}
}

func (p *Provider) WithUpdate(iden *identity.LDAP, loginUserName *string, claims map[string]interface{}, rawEntryJSON map[string]interface{}) *identity.LDAP {
	newIden := *iden
	newIden.Claims = claims
	newIden.RawEntryJSON = rawEntryJSON
	newIden.LastLoginUserName = loginUserName
	return &newIden
}

func (p *Provider) Create(ctx context.Context, i *identity.LDAP) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(ctx, i)
}

func (p *Provider) Update(ctx context.Context, i *identity.LDAP) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(ctx, i)
}

func (p *Provider) Delete(ctx context.Context, i *identity.LDAP) error {
	return p.Store.Delete(ctx, i)
}

func sortIdentities(is []*identity.LDAP) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}

func (p *Provider) MakeSpecFromEntry(ctx context.Context, serverConfig *config.LDAPServerConfig, loginUserName string, entry *ldap.Entry) (*identity.Spec, error) {
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

	err := p.StandardAttributesNormalizer.Normalize(ctx, claims)
	if err != nil {
		return nil, err
	}

	return &identity.Spec{
		Type: model.IdentityTypeLDAP,
		LDAP: &identity.LDAPSpec{
			ServerName:           serverConfig.Name,
			UserIDAttributeName:  userIDAttributeName,
			UserIDAttributeValue: userIDAttributeValue,
			Claims:               claims,
			RawEntryJSON:         entry.ToJSON(),
			LastLoginUserName:    &loginUserName,
		},
	}, nil
}
