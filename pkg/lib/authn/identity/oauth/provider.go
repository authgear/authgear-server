package oauth

import (
	"context"
	"sort"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

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

func (p *Provider) List(ctx context.Context, userID string) ([]*identity.OAuth, error) {
	is, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(ctx context.Context, name string, value string) ([]*identity.OAuth, error) {
	is, err := p.Store.ListByClaim(ctx, name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(ctx context.Context, userID, id string) (*identity.OAuth, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetByProviderSubject(ctx context.Context, providerID oauthrelyingparty.ProviderID, subjectID string) (*identity.OAuth, error) {
	return p.Store.GetByProviderSubject(ctx, providerID, subjectID)
}

func (p *Provider) GetByUserProvider(ctx context.Context, userID string, providerID oauthrelyingparty.ProviderID) (*identity.OAuth, error) {
	return p.Store.GetByUserProvider(ctx, userID, providerID)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*identity.OAuth, error) {
	return p.Store.GetMany(ctx, ids)
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

func (p *Provider) Create(ctx context.Context, i *identity.OAuth) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(ctx, i)
}

func (p *Provider) Update(ctx context.Context, i *identity.OAuth) error {
	now := p.Clock.NowUTC()
	i.UpdatedAt = now
	return p.Store.Update(ctx, i)
}

func (p *Provider) Delete(ctx context.Context, i *identity.OAuth) error {
	return p.Store.Delete(ctx, i)
}

func sortIdentities(is []*identity.OAuth) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
