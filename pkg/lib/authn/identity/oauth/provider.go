package oauth

import (
	"sort"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store          *Store
	Clock          clock.Clock
	IdentityConfig *config.IdentityConfig
}

func (p *Provider) List(userID string) ([]*identity.OAuth, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(name string, value string) ([]*identity.OAuth, error) {
	is, err := p.Store.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.OAuth, error) {
	is, err := p.Store.ListByClaimJSONPointer(pointer, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(userID, id string) (*identity.OAuth, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByProviderSubject(providerID oauthrelyingparty.ProviderID, subjectID string) (*identity.OAuth, error) {
	return p.Store.GetByProviderSubject(providerID, subjectID)
}

func (p *Provider) GetByUserProvider(userID string, providerID oauthrelyingparty.ProviderID) (*identity.OAuth, error) {
	return p.Store.GetByUserProvider(userID, providerID)
}

func (p *Provider) GetMany(ids []string) ([]*identity.OAuth, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) New(
	userID string,
	providerID oauthrelyingparty.ProviderID,
	subjectID string,
	profile map[string]interface{},
	claims map[string]interface{},
) *identity.OAuth {
	i := &identity.OAuth{
		ID:                uuid.New(),
		UserID:            userID,
		ProviderID:        providerID,
		ProviderSubjectID: subjectID,
		UserProfile:       profile,
		Claims:            claims,
	}

	alias := ""
	for _, providerConfig := range p.IdentityConfig.OAuth.Providers {
		providerID := providerConfig.AsProviderConfig().ProviderID()
		if providerID.Equal(i.ProviderID) {
			alias = providerConfig.Alias()
		}
	}
	if alias != "" {
		i.ProviderAlias = alias
	}

	return i
}

func (p *Provider) WithUpdate(
	iden *identity.OAuth,
	rawProfile map[string]interface{},
	claims map[string]interface{},
) *identity.OAuth {
	newIden := *iden
	newIden.UserProfile = rawProfile
	newIden.Claims = claims

	return &newIden
}

func (p *Provider) Create(i *identity.OAuth) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Update(i *identity.OAuth) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(i)
}

func (p *Provider) Delete(i *identity.OAuth) error {
	return p.Store.Delete(i)
}

func sortIdentities(is []*identity.OAuth) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
