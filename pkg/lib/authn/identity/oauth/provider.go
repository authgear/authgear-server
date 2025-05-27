package oauth

import (
	"context"
	"sort"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
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
	spec *identity.OAuthSpec,
) *identity.OAuth {
	i := &identity.OAuth{
		ID:                uuid.New(),
		UserID:            userID,
		ProviderID:        spec.ProviderID,
		ProviderSubjectID: spec.SubjectID,
		UserProfile:       spec.RawProfile,
		Claims:            spec.StandardClaims,
		ProviderAlias:     spec.ProviderAlias,
	}

	if spec.DoNotStoreIdentityAttributes {
		p.stripPII(i)
	}

	return i
}

func (p *Provider) WithUpdate(
	iden *identity.OAuth,
	spec *identity.OAuthSpec,
) *identity.OAuth {
	newIden := *iden
	// For non-Apple provider, we can just update.
	// For Apple, we need to merge given_name and family_name because
	// they only available at THE FIRST TIME authorization.
	if newIden.ProviderID.Type == liboauthrelyingparty.TypeApple {
		newIden.Apple_MergeRawProfileAndClaims(spec.RawProfile, spec.StandardClaims)
	} else {
		newIden.UserProfile = spec.RawProfile
		newIden.Claims = spec.StandardClaims
	}

	if spec.DoNotStoreIdentityAttributes {
		p.stripPII(&newIden)
	}

	return &newIden
}

// stripPII mutates i in-place.
func (p *Provider) stripPII(i *identity.OAuth) {
	// Strip and replace it with an empty map.
	i.UserProfile = make(map[string]any)
	i.Claims = make(map[string]any)
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
